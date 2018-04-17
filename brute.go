package brute

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/jimen0/resolver"
)

// Bruter represents the bruteforcing configuration.
type Bruter struct {
	// Domain is the DNS name for which subdomains are bruteforced.
	Domain string
	// Record is the DNS record that is queried.
	Record string
	// Retries is the number of retries done for each subdomain resolution.
	Retries int
	// Servers is the list of DNS servers that are used.
	Servers []string
	// Workers represents the number of concurrent resolutions that are done.
	Workers int
}

// Brute bruteforces subdomains for a given domain using subdomain names readed from r.
func (b Bruter) Brute(ctx context.Context, r io.Reader, out chan<- string) error {
	defer close(out)

	if isWildcard(ctx, b.Domain, b.Servers[0]) {
		return errors.New("wildcard domain")
	}

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(b.Workers)

	chans := make([]chan string, b.Workers)
	for i := 0; i < len(chans); i++ {
		chans[i] = make(chan string)
	}

	for _, ch := range chans {
		go func(c chan string) {
			defer wg.Done()

			select {
			case <-done:
				return
			default: // avoid blocking
			}

			for v := range c {
				dom := fmt.Sprintf("%v.%s", v, b.Domain)
				dst, err := resolver.Resolve(ctx, b.Record, dom, b.Retries, b.Servers)
				if err != nil {
					continue
				}

				if len(dst) > 0 {
					out <- v
				}
			}
		}(ch)
	}

	var current int
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				done <- struct{}{}
				return err
			}
		default: // avoid blocking.
		}

		val := sc.Text()
		aux := val

		chans[current%b.Workers] <- aux
		current++
	}

	if err := sc.Err(); err != nil && err != io.EOF {
		done <- struct{}{}
		return err
	}

	for _, c := range chans {
		close(c)
	}

	wg.Wait()
	return nil
}

func isWildcard(ctx context.Context, d string, srv string) bool {
	attempts := make([]string, 4)
	for i := 0; i < 4; i++ {
		uuid, err := uuid()
		if err != nil {
			return false
		}
		attempts[i] = uuid
	}

	for _, v := range attempts {
		dom := fmt.Sprintf("%s.%s", v, d)

		dst, err := resolver.Resolve(ctx, "CNAME", dom, 0, []string{srv})
		if err != nil {
			continue
		}
		if len(dst) > 0 {
			return true
		}

		dst, err = resolver.Resolve(ctx, "A", dom, 0, []string{srv})
		if err != nil {
			continue
		}
		if len(dst) > 0 {
			return true
		}
	}

	return false
}

func uuid() (string, error) {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", fmt.Errorf("could not read random UUID: %v", err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
