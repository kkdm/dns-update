package main

import (
    "os"
    "errors"
    "os/signal"
    "syscall"
    "log"
    "github.com/google/logger"
    "github.com/jessevdk/go-flags"

    nats "github.com/nats-io/nats.go"
    "github.com/nats-io/stan.go"
)

var opts Opts
var parser = flags.NewParser(&opts, flags.Default)

func checkEnv() (bool, error) {
    if os.Getenv("READ_TOKEN") == "" || os.Getenv("EDIT_TOKEN") == "" {
        return false, errors.New("test")
    }

    return true, nil
}

func main() {
    if _, err := parser.Parse(); err != nil {
        os.Exit(1)
    }

    logfile, _ := os.OpenFile(opts.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
    defer logfile.Close()

    logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    defer logger.Init("Logger", opts.Verbose, true, logfile).Close()

    if _, err := checkEnv() ; err != nil {
        logger.Fatal("READ_TOKEN and EDIT_TOKEN are required to be set.")
    }

    logger.Infof("server params: server: %s, cluster: %s, log-path: %s, verbose: %t",
        opts.NatsServer, opts.Cluster, opts.LogPath, opts.Verbose)

    nc, err := nats.Connect(opts.NatsServer)
    if err != nil {
        logger.Fatal(err)
    }
    defer nc.Close()

    sc, err := stan.Connect(
        opts.Cluster,
        opts.ClientId,
        stan.NatsConn(nc),
        stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
            logger.Fatalf("Connection lost, reason: %v", reason)
        }))
    if err != nil {
        logger.Fatalf("connection failed: %v, NATS streaming server URL: %v", err, opts.NatsServer)
    }
    logger.Infof("connected. NatsServer: %v, Cluster: [%v], ClientId: [%v]",
        opts.NatsServer,
        opts.Cluster,
        opts.ClientId,
    )

    sub, err := sc.Subscribe(
        opts.Subject,
        postUpdate,
        stan.StartWithLastReceived(),
    )
    if err != nil {
        sc.Close()
        logger.Fatal(err)
    }

    logger.Infof("listening, subject: [%s], clientId: [%s]",
        opts.Subject,
        opts.ClientId,
    )

    sigChan := make(chan os.Signal, 1)
    cleanupDone := make(chan bool)
    signal.Notify(sigChan,
        syscall.SIGHUP,
        syscall.SIGINT,
        syscall.SIGTERM,
        syscall.SIGKILL,
    )
    go func() {
        s := <-sigChan
        logger.Errorf("signal received: %v", s)
        sub.Unsubscribe()
        sc.Close()
        cleanupDone <- true
    }()
    <-cleanupDone
}
