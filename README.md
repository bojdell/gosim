# Gosim: Simulation testing for Go

Gosim is a simulation testing framework that aims to make it easier to write
reliable distributed systems code in go. Gosim takes standard go programs and
tests and runs them with a simulated network, filesystem, multiple machines, and
more.

In Gosim's simulated environment you can run an entire distributed system in a
single go process, without having to set up any real servers. The simulation can
also introduce all kinds of problems, such as flaky networks, restarting
machines, or lost writes on a filesystem, to test that the program survive those
problems before they happen in real life.

An interesting and entertaining introduction to simulation testing is the talk
[Testing Distributed Systems w/ Deterministic Simulation from Strangeloop
2014](https://www.youtube.com/watch?v=4fFDFbi3toc).

Gosim is an experimental project. Feedback, suggestions, and ideas are all very
welcome. The underlying design feels quite solid, but the implemented and
simulated APIs are still limited: files work, but not directories; TCP works,
but not UDP; IP addresses work, but not hostnames. The API can and will change.
Now that those warnings are out of the way, please take a look at what gosim
can do.

# Using Gosim

## From go test to gosim test
Gosim tests standard go code and is used just like another go package. To get
started with `gosim`, you need to import it in your module:
```
> mkdir example && cd example
> go mod init example
> go get github.com/jellevandenhooff/gosim
```
To test code with Gosim, write a small a test in a file `simple_test.go` and then run
it using `go test`:
```go
package example_test

import (
	"testing"

	"github.com/jellevandenhooff/gosim"
)

func TestGosim(t *testing.T) {
	t.Logf("Are we in the Matrix? %v", gosim.IsSim())
}
```
```
> go test -v -run TestGosim
=== RUN   TestGosim
    simple_test.go:10: Are we in the Matrix? false
--- PASS: TestGosim (0.00s)
PASS
ok  	example	0.216s
```
To run this test with `gosim` instead, replace `go test` with
`go run github.com/jellevandenhooff/gosim/cmd/gosim test`:
```
> go run github.com/jellevandenhooff/gosim/cmd/gosim test -v -run TestGosim
=== RUN   TestGosim
    1 main/4     14:10:03.000 INF example/simple_gosim_test.go:11 > Are we in the Matrix? true method=t.Logf
--- PASS: TestGosim (0.00s)
ok  	translated/example	0.204s
```
The `gosim test` command has flags similar to `go test`. The test output is
more involved than a normal `go test`. Every log line includes the simulated
machine and the goroutine that invoked the log to help debug tests running on
multiple machines.

If running gosim fails with errors about missing `go.sum` entries, run
`go mod tidy` to update your `go.sum` file.

## Simulation
Tests running in Gosim run inside Gosim's simulation environment. In the
simulation, tests can create simulated machines that can talk to each other
over a simulated network, crash and restart machine, introduce latency,
and more. The following longer example shows some of those features:
```go
package example_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/jellevandenhooff/gosim"
)

var count = 0

func server() {
	log.Printf("starting server")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		count++
		log.Printf("got a request from %s", r.RemoteAddr)
		fmt.Fprintf(w, "hello from the server! request: %d", count)
	})
	http.ListenAndServe("10.0.0.1:80", nil)
}

func request() {
	log.Println("making a request")
	resp, err := http.Get("http://10.0.0.1/")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(body))
}

func TestMachines(t *testing.T) {
	// run the server
	serverMachine := gosim.NewMachine(gosim.MachineOpts{
		Label:    "server",
		Addr:     netip.MustParseAddr("10.0.0.1"),
		MainFunc: server,
	})

	// let the server start
	time.Sleep(time.Second)

	// make some requests to see them work
	request()
	request()

	// restart the server
	log.Println("restarting the server")
	serverMachine.Crash()
	serverMachine.Restart()
	time.Sleep(time.Second)

	// make a new request to see a reset count
	request()

	// add some latency
	log.Println("adding latency")
	gosim.SetDelay("10.0.0.1", "11.0.0.1", time.Second)

	// make another request to see the latency
	request()
}
```
This example contains a HTTP server running on its own gosim machine, and a
client that makes several requests to the server. Then the client
messes with the server, restarting it and adding latency. Let's see
what happens this test is run:
<!-- TODO: two machines and a bug? -->
```
> go run github.com/jellevandenhooff/gosim/cmd/gosim test -v -run TestMachines
=== RUN   TestMachines
    1 server/5   14:10:03.000 INF example/machines_gosim_test.go:17 > starting server
    2 main/4     14:10:04.000 INF example/machines_gosim_test.go:27 > making a request
    3 server/8   14:10:04.000 INF example/machines_gosim_test.go:20 > got a request from 11.0.0.1:10000
    4 main/4     14:10:04.000 INF example/machines_gosim_test.go:41 > hello from the server! request: 1
    5 main/4     14:10:04.000 INF example/machines_gosim_test.go:27 > making a request
    6 server/8   14:10:04.000 INF example/machines_gosim_test.go:20 > got a request from 11.0.0.1:10000
    7 main/4     14:10:04.000 INF example/machines_gosim_test.go:41 > hello from the server! request: 2
    8 main/4     14:10:04.000 INF example/machines_gosim_test.go:60 > restarting the server
    9 server/13  14:10:04.000 INF example/machines_gosim_test.go:17 > starting server
   10 main/4     14:10:05.000 INF example/machines_gosim_test.go:27 > making a request
   11 server/16  14:10:05.000 INF example/machines_gosim_test.go:20 > got a request from 11.0.0.1:10001
   12 main/4     14:10:05.000 INF example/machines_gosim_test.go:41 > hello from the server! request: 1
   13 main/4     14:10:05.000 INF example/machines_gosim_test.go:69 > adding latency
   14 main/4     14:10:05.000 INF example/machines_gosim_test.go:27 > making a request
   15 server/16  14:10:06.000 INF example/machines_gosim_test.go:20 > got a request from 11.0.0.1:10001
   16 main/4     14:10:07.000 INF example/machines_gosim_test.go:41 > hello from the server! request: 2
--- PASS: TestMachines (4.00s)
ok  	translated/example	0.237s
```
In the log, each line is annotated with the machine and goroutine that wrote
the log. When debugging code running on multiple machines, it is helpful to
know where logs are coming from.

The first thing that happens is the server starting at step 1. Then the client
makes a request at step 2, the server receives it, responds, and the client
prints the response. When the client makes another request the server this
time responds with a higher count at steps 6 and 7.

Afterwards, the client restarts the server at step 8 and the server prints that
is starting again at step 9. When the client makes a request after the restart
the server responds once again with a count of 1 at steps 11 and 12.  How did
that happen? After the server is restarted, the counter resets to zero. Each
machine has its own copy of global variables which get re-initialized when a
machine crashes.

Notice that the client address (after "got a request from:" in the logs) also
changes after the restart, and the goroutine handling the requests on the server
changes. The simulated TCP connection breaks when the machine crashes and the
`net/http` client makes a new connection afterwards.

After the restart, the client calls a Gosim API to add latency to the simulated network
at step 13. The earlier requests all got sent and received at the same timestamp,
such as on steps 10, 11, and 12 all at 14:10:05. The request after adding
latency is sent at 14:10:05 on step 14 and received a second later at 14:10:06
on step 15. This is the added latency in action.

Even though the test takes over 4 simulated seconds, running the tests takes
less than a second. Simulated time can be sped up when Gosim knows that all
goroutines are waiting for time advance. With simulated time, tests can use
realistic latency and timeouts without development time becoming slow.

## Debugging
Gosim's simulation is deterministic, which means that running a test twice
will result in the exact same behavior, from randomly generated numbers
to goroutine concurrency interleavings. That is cool because it means that
if we see an interesting log after running the program but do not fully
understand why that log printed a certain value, we can re-run the program
and see exactly what happened at the moment of printing.

The numbers at the start of each log are Gosim's step numbers. Take step 11 from
the log above, from the line "got a request ...". We can debug the state of the
program at the time the log was printed with `gosim debug`:
```
> go run github.com/jellevandenhooff/gosim/cmd/gosim debug -package=. -test=TestMachines -step=11
...
    26:	func (w gosimSlogHandler) Handle(ctx context.Context, r slog.Record) error {
    27:		r.AddAttrs(slog.Int("goroutine", gosimruntime.GetGoroutine()))
=>  28:		r.AddAttrs(slog.Int("step", gosimruntime.Step()))
    29:		return w.inner.Handle(ctx, r)
    30:	}
...
```
The `gosim debug` command runs the requested test inside of the Delve go
debugger and stops the test at the requested step. The debugger here puts us in
Gosim's `log/slog` handler which calls `gosimruntime.Step()` and includes the
step number.

Now we can debug the program as a standard Go program. To see what was happening
in the HTTP handler that called print, we run backtrace
```
(dlv) bt
 0  0x000000010129e7ac in translated/github.com/jellevandenhooff/gosim/internal_/simulation.gosimSlogHandler.Handle
    at ./github.com/jellevandenhooff/gosim/internal_/simulation/userspace_gosim.go:28
 1  0x00000001012a5e24 in translated/github.com/jellevandenhooff/gosim/internal_/simulation.(*gosimSlogHandler).Handle
    at <autogenerated>:1
 2  0x0000000101283d68 in translated/log/slog.(*handlerWriter).Write
    at ./log/slog/logger_gosim.go:99
 3  0x000000010124a600 in translated/log.(*Logger).output
    at ./log/log_gosim.go:242
 4  0x000000010124ae38 in translated/log.Printf
    at ./log/log_gosim.go:394
 5  0x000000010160e03c in translated/example_test.server.func1
    at ./example/machines_gosim_test.go:20
...
```
and then select the HTTP handler's stack frame that called `log.Printf`
```
(dlv) frame 5
> translated/github.com/jellevandenhooff/gosim/internal_/simulation.gosimSlogHandler.Handle() ./github.com/jellevandenhooff/gosim/internal_/simulation/userspace_gosim.go:28 (PC: 0x1029ce45c)
Frame 5: ./example/machines_gosim_test.go:19 (PC: 102d3decc)
    14:	)
    15:
    16:	func server() {
    17:		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    18:			G().count++
=>  19:			log.Printf("got a request from %s", r.RemoteAddr)
    20:			fmt.Fprintf(w, "hello from the server! request: %d", G().count)
    21:		})
    22:		http.ListenAndServe("10.0.0.1:80", nil)
    23:	}
    24:
```
Now we can inspect the state of memory to see details that the log line might have missed
```
(dlv) print r.RemoteAddr
"11.0.0.1:10001"
```

# API and documentation

A description of Gosim's architecture and design decisions is in
[docs/design.md](docs/design.md).

Gosim's public packages are:

- [github.com/jellevandenhooff/gosim](./): the core API for creating machines
  and manipulating the simulation environment

- [github.com/jellevandenhooff/gosim/cmd/gosim](./cmd/gosim): the CLI for
  running Gosim tests

- [github.com/jellevandenhooff/gosim/metatesting](./metatesting/): a package
  for running Gosim tests inside of normal go test.

- [github.com/jellevandenhooff/gosim/nemesis](./nemesis/): a package, still
  sparse, to introduce chaos into simulations

# Development

Gosim and its tools are tested using the `./test.sh` script, which invokes the
tests defined in `./Taskfile.yml`. These test the Gosim tooling as well as the
behavior of simulated code. Ideally tests verify that the simulation behaves
like reality: The tests for the filesystem in
[github.com/jellevandenhooff/gosim/internal/tests/behavior/disk_test.go](./internal/tests/behavior/disk_test.go)
are run in both simulated and non-simulated builds.

Gosim simulates Linux, but should work on macOS, Linux, and Windows (not tested
on Windows) running on either arm64 or amd64. To test that code works on Linux
amd64 and arm64, there are scripts `.ci/crossarch-tests/test-amd64.sh` and
`.ci/crossarch-tests/test-arm64.sh` that run the tests using Docker. A Github
Action runs the tests on Linux amd64.
