package dregistry

import "github.com/streamingfast/logging"

var zlogTest, _ = logging.PackageLogger("dregistry_test", "github.com/streamingfast/dregistry(test)")

func init() { logging.InstantiateLoggers() }
