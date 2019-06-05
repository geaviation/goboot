package util

import (
	"errors"
	"testing"
	"time"
)

func TestRetryDefaultBackOff(t *testing.T) {
	//t.Skipf("skipping...")

	const retryCount = 3 //default is 3
	var i = 0

	err := Retry(func() error {
		i++
		log.Printf("function is called: %d time\n", i)

		if i == retryCount {
			log.Println("OK")
			return nil
		}

		log.Println("faking error")
		return errors.New("some error")
	})
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if i != retryCount {
		t.Errorf("Invalid number of retries: %d", i)
	}
}

func TestRetryBackOffNoRetry(t *testing.T) {
	//t.Skipf("skipping...")

	const retryCount = 1
	var i = 0

	err := Retry(func() error {
		i++
		log.Printf("function is called: %d time\n", i)

		if i == retryCount {
			log.Println("OK")
			return nil
		}

		log.Println("faking error")
		return errors.New("some error")
	}, NewDefaultBackOff())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if i != retryCount {
		t.Errorf("Invalid number of retries: %d", i)
	}
}

func TestRetryBackOff(t *testing.T) {
	//t.Skipf("skipping...")

	const retryCount = 9
	const duration = time.Millisecond * 3
	var i = 0

	err := Retry(func() error {
		i++
		log.Printf("function is called: %d time\n", i)

		if i == retryCount {
			log.Println("OK")
			return nil
		}

		log.Println("faking error")
		return errors.New("some error")
	}, NewBackOff(retryCount, duration))
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if i != retryCount {
		t.Errorf("Invalid number of retries: %d", i)
	}
}
