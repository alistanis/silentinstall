# silentinstall

Expect - but simpler! [Linux Expect Man Page](https://linux.die.net/man/1/expect)

Ever wanted to automate installing loud linux/unix packages or scripts easily? Now you can. Hell, it might even work for Windows, but I haven't tried.

SilentInstall is a package that handles running your commands for you, and handles their expected inputs/outputs in a way you define.

Get it: `go get -u github.com/alistanis/silentinstall`

# What is this for? Doesn't everything run in Docker or have an automated install now?

Unfortunately, for many of us, especially at Universities, this isn't the case. There's lots of legacy cruft lying around that's now finding its way into the cloud, and that poses problems for automation.

# Is this a replacement for Expect?

No. It is not a DSL or Scripting language and it is not as fully featured as expect is. It is much simpler, however, and the configuration format is easy to understand.

# Config Format

Below, we'll see a basic json structure that accepts a list of SilentCmds as objects.
The "cmd" parameter is a string, and the {{.GOPATH}} will be interpreted as an environment variable. This can be done for any variable currently in your environment following the Go Template format.
The "io" parameter is a list of "input" and "output" objects. You specify what the input is, what the output should be, and silentinstall handles the rest.
In order to specify a newline, just leave an empty string in "output".
```
    [
      {
        "cmd": "{{.GOPATH}}/src/github.com/alistanis/silentinstall/silent/test_data/multiple_io.sh",
        "io": [
          {
            "input": "Hello! Please enter your name!", "output": "Chris"
          },
          {
            "input": "Please enter your age!", "output": "29"
          }
        ]
      }
    ]
```

# Running SilentInstall

```
    silentinstall -f silent/test_data/multiple_io_example_config.json
    
    [/Users/cmc666/work/polaris/src/github.com/alistanis/silentinstall/silent/test_data/multiple_io.sh]
    2016/12/01 14:52:36 ui.go:231: ui: Hello! Please enter your name!
    
    2016/12/01 14:52:36 ui.go:231: ui: Chris
    2016/12/01 14:52:36 ui.go:231: ui: Please enter your age!
    
    2016/12/01 14:52:36 ui.go:231: ui: 29
    2016/12/01 14:52:36 ui.go:231: ui: SilentInstall has finished successfully!
```

# Usage

```
    Usage of ./silentinstall:
      -f string
        	The path of the config file
      -file string
        	The path of the config file
      -v	Prints verbose output if true
```

# Running the Tests

All test dependencies are included in the vendor folder. Simply cd to the source directory and run go test -v
```
    cd $GOPATH/src/github.com/alistanis/silentinstall/silent
    go test -v
```

# Special Thanks

The guys over at SmartyStreets for [Goconvey](http://goconvey.co/), which I use in all my projects, jtolds for his goroutine local storage package (which I don't use but Goconvey does) https://github.com/jtolds/gls, and Mitchell Hashimoto and the guys at [Hashicorp](https://www.hashicorp.com/).