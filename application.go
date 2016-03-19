package glaze

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// Application encapsulated all the startup logic for a web app using glaze.
type Application struct {
	port        int
	graceful    bool
	verbose     bool
	listener    gracefulListener
	server      *http.Server
	processName string
	accessLog   *os.File
	errorLog    *os.File
}

// Initialize handles all the pre-launch initialization tasks.
func (app *Application) Initialize(processName string, port int, verbose, graceful bool) error {
	app.port = port
	app.verbose = verbose
	app.graceful = graceful

	app.processName = processName

	err := app.parseConfig()
	if nil != err {
		return err
	}

	logPath := viper.GetString("LogPath")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		err = os.Mkdir(logPath, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Set up server and listener
	app.server = &http.Server{
		Addr:           fmt.Sprintf(":%d", app.port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 16,
	}

	var l net.Listener

	if app.graceful {
		f := os.NewFile(3, "")
		l, err = net.FileListener(f)
	} else {
		l, err = net.Listen("tcp", app.server.Addr)
	}
	if err != nil {
		return err
	}

	if app.graceful {
		parent := syscall.Getppid()
		log.Printf("main: Killing parent pid: %v", parent)
		syscall.Kill(parent, syscall.SIGTERM)
	}

	app.listener = gracefulListener{Listener: l, stop: make(chan signal, 1)}

	// Configure the access and error logs
	log.SetPrefix(fmt.Sprintf("[%5d] ", syscall.Getpid()))

	app.accessLog, err = app.CreateLog("access.log")
	if err != nil {
		return err
	}

	app.errorLog, err = app.CreateLog("error.log")
	if err != nil {
		return err
	}

	return nil
}

func (app Application) parseConfig() error {
	// Create configuration
	viper.Set("AppRoot", os.ExpandEnv("."))
	viper.Set("LogPath", os.ExpandEnv("./logs/"))

	viper.SetConfigName("config")
	viper.AddConfigPath(os.ExpandEnv("./config/"))
	err := viper.ReadInConfig()

	if err != nil {
		return err
	}

	viper.Set("FullViewPath", path.Join(viper.GetString("AppRoot"), viper.GetString("ViewPath")))

	return nil
}

func (app Application) logreq(req *http.Request) {
	if app.verbose {
		log.Printf("%v %v from %v", req.Method, req.URL, req.RemoteAddr)
	}
}

// Restart performs a graceful restart of the process
func (app Application) Restart(w http.ResponseWriter, req *http.Request) {
	file := app.listener.File() // this returns a Dup()

	cmd := exec.Command(fmt.Sprintf("./%s", app.processName),
		fmt.Sprintf("-port=%d", app.port),
		fmt.Sprintf("-verbose=%v", app.verbose),
		"-graceful")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = []*os.File{file}

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Graceful restart: Failed to launch, error: %v", err)
	}
}

// CreateLog either opens or creates and opens a log file at the given path, and returns an append handle.
func (app Application) CreateLog(name string) (*os.File, error) {
	fullPath := path.Join(viper.GetString("LogPath"), name)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		_, err := os.Create(fullPath)
		if err != nil {
			return nil, err
		}
	}

	handle, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0600)

	if err != nil {
		return nil, err
	}

	return handle, nil
}

// ServeWithAutoRouting computes standard paths to every Controller that
// has been added to the Application, and then starts the application server.
func (app Application) ServeWithAutoRouting(controllers []Controller) error {
	// Create routes from the controllers
	autoRouter := mux.NewRouter()

	fmt.Println("Generating routes...")

	for _, controller := range controllers {
		for path, handler := range GenerateRoutes(controller) {
			fmt.Printf("\t%s\n", path)
			if path == "/public/index" {
				autoRouter.HandleFunc("/", handler)
			} else {
				autoRouter.HandleFunc(path, handler)
			}
		}
	}

	// Tell the router how to handle static assets
	autoRouter.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("views/static"))))

	err := app.ServeWithHandler(autoRouter)
	if err != nil {
		return err
	}

	return nil
}

// ServeWithHandler starts the server with a client-provided handler.
// If you want to use glaze's auto-routing, use ServeWithAutoRouting instead.
func (app Application) ServeWithHandler(handler http.Handler) error {
	// Wrap the router in a LoggingHandler - this will use accessLog to log every connection to the server.
	loggingRouter := handlers.LoggingHandler(app.accessLog, handler)

	// When we hit the /upgrade URL, perform a graceful restart.
	// TODO - this needs to change, but it's fine for proof-of-concept
	http.HandleFunc("/upgrade", app.Restart)

	// For any other connections, use the loggingRouter
	http.Handle("/", loggingRouter)

	// this goroutine monitors the channel. Can't do this in
	// Accept (below) because once it enters listener.Listener.Accept()
	// it blocks. We unblock it by closing the fd it is trying to
	// accept(2) on.
	go func() {
		_ = <-app.listener.stop
		app.listener.stopped = true
		app.listener.Listener.Close()
	}()

	fmt.Println("Serving...")

	err := app.server.Serve(app.listener)
	if err != nil {
		return err
	}

	return nil
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	http.Error(w, "Could not find the requested page", http.StatusInternalServerError)
}
