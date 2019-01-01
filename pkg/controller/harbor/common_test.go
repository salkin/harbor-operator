package harbor

import (
	"fmt"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func initLog() {
	logf.SetLogger(&Logg{})
}

type Logg struct {
}

func (l *Logg) V(level int) logr.InfoLogger {
	return l

}

func (l *Logg) WithValues(keysAndValues ...interface{}) logr.Logger {
	return l

}

func (l *Logg) WithName(name string) logr.Logger {
	return l

}

func (l *Logg) Error(err error, msg string, keysAndValues ...interface{}) {

	fmt.Printf("Err %v: Msg %s", err, msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		fmt.Printf("%s:%v", keysAndValues[i], keysAndValues[i+1])
	}
}

func (l *Logg) Info(msg string, keysAndValues ...interface{}) {
	fmt.Printf(msg)
	for i := 0; i < len(keysAndValues); i += 2 {
		fmt.Printf("%s:%v", keysAndValues[i], keysAndValues[i+1])
	}

}

func (l *Logg) Enabled() bool {
	panic("not implemented")

}
