package driver

import (
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

func TestMakeCSIPanicRecovery_Recovers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	log := logger.Sugar()
	var err error
	func() {
		defer func() { log.Infof("Error returned from panic recory : %v", err) }()
		defer MakeCSIPanicRecoveryWithError(log, nil, "UnitTestOp", map[string]string{"test": "true"}, &err, codes.Internal)()
		panic("boom")
		//code after panic doesn't run, but this as we are recovering from panic, test suite will not crash
		//which will validate this functionality
	}()

}
