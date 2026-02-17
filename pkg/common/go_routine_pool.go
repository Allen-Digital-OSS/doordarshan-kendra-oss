package common

import (
	"github.com/google/wire"
	"github.com/labstack/gommon/log"
	"github.com/panjf2000/ants/v2"
	"time"
)

var ProviderSet = wire.NewSet(NewGoRoutinePool)

type GoRoutinePool struct {
	appConfig *AppConfig
	pool      *ants.Pool
}

// NewGoRoutinePool creates a new goroutine pool with the given pool size.
func NewGoRoutinePool(appConfig *AppConfig) (*GoRoutinePool, error) {
	poolSize := appConfig.AppParams.GoRoutinePoolSize
	antsPool, err := ants.NewPool(poolSize)
	if err != nil {
		return nil, err
	}
	return &GoRoutinePool{
		appConfig: appConfig,
		pool:      antsPool,
	}, nil
}

// Go submits a task to the pool.
func (grp *GoRoutinePool) Go(task func()) error {
	log.Info("Starting a go routine in the pool")
	err := grp.pool.Submit(task)
	log.Info("Pool size: ", grp.pool.Cap())
	log.Info("Running size: ", grp.pool.Running())
	log.Info("Free size: ", grp.pool.Free())
	if err != nil {
		log.Error("Error starting a go routine in the pool: ", err)
	}
	return err
}

// Release releases the pool.
func (grp *GoRoutinePool) Release() {
	grp.pool.Release()
}

// GetPoolSize returns the pool size.
func (grp *GoRoutinePool) GetPoolSize() int {
	return grp.pool.Cap()
}

// GetRunningSize returns the running size.
func (grp *GoRoutinePool) GetRunningSize() int {
	return grp.pool.Running()
}

// GetFreeSize returns the free size.
func (grp *GoRoutinePool) GetFreeSize() int {
	return grp.pool.Free()
}

// IsClosed returns true if the pool is closed.
func (grp *GoRoutinePool) IsClosed() bool {
	return grp.pool.IsClosed()
}

// Tune changes the pool size.
func (grp *GoRoutinePool) Tune(poolSize int) {
	grp.pool.Tune(poolSize)
}

// ReleaseTimeout releases the pool with timeout.
func (grp *GoRoutinePool) ReleaseTimeout(timeout time.Duration) error {
	err := grp.pool.ReleaseTimeout(timeout)
	return err
}
