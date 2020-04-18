package pool

import (
	"errors"
	"testing"
	"time"
)

func TestPoolRuns(t *testing.T) {

	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
	}

	p := NewPool(tasks, 3)
	p.Run()

	var numErrors int
	for _, task := range p.Tasks {
		if task.Err != nil {
			t.Errorf(task.Err.Error())
			numErrors++
		}
		if numErrors >= 10 {
			t.Errorf("Too many errors.")
			break
		}
	}

}

func TestPoolCommunicates(t *testing.T) {

	c := make(chan int, 10)

	tasks := []*Task{
		NewTask(func() error {
			c <- 1
			return nil
		}),
		NewTask(func() error {
			c <- 2
			return nil
		}),
		NewTask(func() error {
			c <- 3
			return nil
		}),
	}

	p := NewPool(tasks, 3)
	p.Run()

	for i := 0; i < 3; i++ {
		select {
		case <-time.After(10 * time.Millisecond):
			t.Errorf("Timeout")
		case <-c:
			continue
		}
	}

	var numErrors int
	for _, task := range p.Tasks {
		if task.Err != nil {
			t.Errorf(task.Err.Error())
			numErrors++
		}
		if numErrors >= 10 {
			t.Errorf("Too many errors.")
			break
		}
	}

}
func TestPoolDetectsError(t *testing.T) {

	tasks := []*Task{
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return nil }),
		NewTask(func() error { return errors.New("expectedError") }),
		NewTask(func() error { return errors.New("expectedError") }),
	}

	p := NewPool(tasks, 3)
	p.Run()

	var numErrors int
	for _, task := range p.Tasks {
		if task.Err != nil {
			numErrors++
		}
	}
	if numErrors != 2 {
		t.Errorf("wrong number of errors wanted 2. got %d", numErrors)
	}

}

func TestPoolLargeTasksList(t *testing.T) {

	N := 500

	c := make(chan int, N)

	tasks := []*Task{}

	for i := 0; i < N; i++ {
		i := i
		newtask := NewTask(func() error {
			c <- i
			return nil
		})
		tasks = append(tasks, newtask)
	}

	p := NewPool(tasks, 32)
	p.Run()

	total := 0
LOOP:
	for i := 0; i < N; i++ {
		select {
		case <-time.After(1 * time.Millisecond):
			t.Errorf("Timeout")
			break LOOP
		case val := <-c:
			total = total + val
		}
	}

	// area of a triangle - checks we did loop iterator reference correctly
	// https://github.com/golang/go/wiki/CommonMistakes
	if total != (N-1)*(N)/2.0 {
		t.Errorf("something's not right with the dispatching %d != %d", total, (N-1)*(N)/2.0)
	}

	var numErrors int
	for _, task := range p.Tasks {
		if task.Err != nil {
			numErrors++
		}
	}
	if numErrors > 0 {
		t.Errorf("got an error")
	}

}
