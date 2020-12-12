package main

type Opts struct {
    NatsServer      string  `short:"s"  long:"server"       description:"nats server url" required:"true"`
    Cluster         string  `short:"c"  long:"cluster"      description:"nats cluster name" required:"true"`
    ClientId        string  `short:"i"  long:"client-id"    description:"nats stream client id" required:"true"`
    LogPath         string  `short:"l"  long:"log-path"     description:"log file path" default:"./server.log"`
    Subject         string  `short:"S"  long:"subject"      description:"subject to subscribe" required:"true"`
    Verbose         bool    `short:"v"  long:"verbose"      description:"verbose log"`
}
