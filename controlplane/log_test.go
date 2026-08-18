package controlplane

import "github.com/streamingfast/logging"

var zlogTest, _ = logging.PackageLogger("dregistry-controlplane_test", "github.com/streamingfast/dregistry/controlplane(test)")

func init() { logging.InstantiateLoggers() }
