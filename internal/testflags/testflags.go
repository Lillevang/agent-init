package testflags

import "flag"

var Update = flag.Bool("update", false, "update golden testdata")
