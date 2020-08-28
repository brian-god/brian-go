package xgo

import (
	"fmt"
	"github.com/brian-god/brian-go/pkg/utils/xstring"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"runtime"
	"sync"
)

// 创建一个迭代器
func SerialUntilError(fns ...func() error) func() error {
	return func() error {
		for _, fn := range fns {
			if err := try(fn, nil); err != nil {
				return err
				// return errors.Wrap(err, xstring.FunctionName(fn))
			}
		}
		return nil
	}
}

// Go goroutine 开启一个goroutine
func Go(fn func()) {
	go try2(fn, nil)
}

func try2(fn func(), cleaner func()) (ret error) {
	if cleaner != nil {
		defer cleaner()
	}
	defer func() {
		_, file, line, _ := runtime.Caller(5)
		if err := recover(); err != nil {
			fmt.Println(file, line)
			//logger.Error("recover", zap.Any("err", err), zap.String("line", fmt.Sprintf("%s:%d", file, line)))
			if _, ok := err.(error); ok {
				ret = err.(error)
			} else {
				ret = fmt.Errorf("%+v", err)
			}
		}
	}()
	fn()
	return nil
}

// ParallelWithError ...
func ParallelWithError(fns ...func() error) func() error {
	return func() error {
		eg := errgroup.Group{}
		for _, fn := range fns {
			eg.Go(fn)
		}

		return eg.Wait()
	}
}
func try(fn func() error, cleaner func()) (ret error) {
	if cleaner != nil {
		defer cleaner()
	}
	defer func() {
		if err := recover(); err != nil {
			_, file, line, _ := runtime.Caller(2)
			logrus.Error("recover", zap.Any("err", err), zap.String("line", fmt.Sprintf("%s:%d", file, line)))
			if _, ok := err.(error); ok {
				ret = err.(error)
			} else {
				ret = fmt.Errorf("%+v", err)
			}
			ret = errors.Wrap(ret, fmt.Sprintf("%s:%d", xstring.FunctionName(fn), line))
		}
	}()
	return fn()
}

// ParallelWithErrorChan calls the passed functions in a goroutine, returns a chan of errors.
// fns会并发执行，chan error
func ParallelWithErrorChan(fns ...func() error) chan error {
	total := len(fns)
	errs := make(chan error, total)

	var wg sync.WaitGroup
	wg.Add(total)

	go func(errs chan error) {
		wg.Wait()
		close(errs)
	}(errs)

	for _, fn := range fns {
		go func(fn func() error, errs chan error) {
			defer wg.Done()
			errs <- try(fn, nil)
		}(fn, errs)
	}

	return errs
}

// RestrictParallelWithErrorChan calls the passed functions in a goroutine, limiting the number of goroutines running at the same time,
// returns a chan of errors.
func RestrictParallelWithErrorChan(concurrency int, fns ...func() error) chan error {
	total := len(fns)
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > total {
		concurrency = total
	}
	var wg sync.WaitGroup
	errs := make(chan error, total)
	jobs := make(chan func() error, concurrency)
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		//consumer
		go func(jobs chan func() error, errs chan error) {
			defer wg.Done()
			for fn := range jobs {
				errs <- try(fn, nil)
			}
		}(jobs, errs)
	}
	go func(errs chan error) {
		//producer
		for _, fn := range fns {
			jobs <- fn
		}
		close(jobs)
		//wait for block errs
		wg.Wait()
		close(errs)
	}(errs)
	return errs
}
