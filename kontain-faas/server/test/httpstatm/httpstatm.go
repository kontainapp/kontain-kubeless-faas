// This work is based on https://github.com/tcnksm/go-httpstat and the license is in LICENSE

package httpstatm

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http/httptrace"
	"strings"
	"time"
)

func min(a, b time.Duration, count int) time.Duration {
	if a <= b && count != 1 {
		return a
	}
	return b
}

// Result stores httpstat info.
type Result struct {
	// CURRENT
	// The following are duration for each phase
	DNSLookup        time.Duration
	TCPConnection    time.Duration
	TLSHandshake     time.Duration
	ServerProcessing time.Duration
	contentTransfer  time.Duration

	// The followings are timeline of request
	NameLookup    time.Duration
	Connect       time.Duration
	Pretransfer   time.Duration
	StartTransfer time.Duration
	total         time.Duration

	// AVERAGE
	// The following are average duration for each phase
	DNSLookupSum        time.Duration
	TCPConnectionSum    time.Duration
	TLSHandshakeSum     time.Duration
	ServerProcessingSum time.Duration
	contentTransferSum  time.Duration

	// The followings are sums timeline of request
	NameLookupSum    time.Duration
	ConnectSum       time.Duration
	PretransferSum   time.Duration
	StartTransferSum time.Duration
	totalSum         time.Duration

	// BEST
	// The following are duration for each phase
	DNSLookupBest        time.Duration
	TCPConnectionBest    time.Duration
	TLSHandshakeBest     time.Duration
	ServerProcessingBest time.Duration
	contentTransferBest  time.Duration

	// The followings are timeline of request
	NameLookupBest    time.Duration
	ConnectBest       time.Duration
	PretransferBest   time.Duration
	StartTransferBest time.Duration
	totalBest         time.Duration

	t0 time.Time
	t1 time.Time
	t2 time.Time
	t3 time.Time
	t4 time.Time
	t5 time.Time // need to be provided from outside

	dnsStart      time.Time
	dnsDone       time.Time
	tcpStart      time.Time
	tcpDone       time.Time
	tlsStart      time.Time
	tlsDone       time.Time
	serverStart   time.Time
	serverDone    time.Time
	transferStart time.Time
	trasferDone   time.Time // need to be provided from outside

	// isTLS is true when connection seems to use TLS
	isTLS bool

	// isReused is true when connection is reused (keep-alive)
	isReused bool

	// Request count
	reqCount int
}

func (r *Result) sums() {
	// The following are average duration for each phase
	r.DNSLookupSum += r.DNSLookup
	r.TCPConnectionSum += r.TCPConnection
	r.TLSHandshakeSum += r.TLSHandshake
	r.ServerProcessingSum += r.ServerProcessing
	r.contentTransferSum += r.contentTransfer

	// The followings are average timeline of request
	r.NameLookupSum += r.NameLookup
	r.ConnectSum += r.Connect
	r.PretransferSum += r.Pretransfer
	r.StartTransferSum += r.StartTransfer
	r.totalSum += r.total
}

func (r *Result) durations() map[string]time.Duration {
	return map[string]time.Duration{
		"DNSLookup":        r.DNSLookup,
		"TCPConnection":    r.TCPConnection,
		"TLSHandshake":     r.TLSHandshake,
		"ServerProcessing": r.ServerProcessing,
		"ContentTransfer":  r.contentTransfer,

		"NameLookup":    r.NameLookup,
		"Connect":       r.Connect,
		"Pretransfer":   r.Connect,
		"StartTransfer": r.StartTransfer,
		"Total":         r.total,
	}
}

func (r *Result) best() {
	// The following are best duration for each phase
	r.DNSLookupBest = min(r.DNSLookupBest, r.DNSLookup, r.reqCount)
	r.TCPConnectionBest = min(r.TCPConnectionBest, r.TCPConnection, r.reqCount)
	r.TLSHandshakeBest = min(r.TLSHandshakeBest, r.TLSHandshake, r.reqCount)
	r.ServerProcessingBest = min(r.ServerProcessingBest, r.ServerProcessing, r.reqCount)
	r.contentTransferBest = min(r.contentTransferBest, r.contentTransfer, r.reqCount)

	// The followings are best timeline of request
	r.NameLookupBest = min(r.NameLookupBest, r.NameLookup, r.reqCount)
	r.ConnectBest = min(r.ConnectBest, r.Connect, r.reqCount)
	r.PretransferBest = min(r.PretransferBest, r.Pretransfer, r.reqCount)
	r.StartTransferBest = min(r.StartTransferBest, r.StartTransfer, r.reqCount)
	r.totalBest = min(r.totalBest, r.total, r.reqCount)
}

// Format formats stats result.
func (r Result) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "                   Average           Best\n")
			fmt.Fprintf(&buf, "DNS lookup:        %4d ms        %4d ms \n",
				int(r.DNSLookupSum/time.Millisecond)/r.reqCount, int(r.DNSLookupBest/time.Millisecond))
			fmt.Fprintf(&buf, "TCP connection:    %4d ms        %4d ms\n",
				int(r.TCPConnectionSum/time.Millisecond)/r.reqCount, int(r.TCPConnectionBest/time.Millisecond))
			fmt.Fprintf(&buf, "TLS handshake:     %4d ms        %4d ms\n",
				int(r.TLSHandshakeSum/time.Millisecond)/r.reqCount, int(r.TLSHandshakeBest/time.Millisecond))
			fmt.Fprintf(&buf, "Server processing: %4d ms        %4d ms\n",
				int(r.ServerProcessingSum/time.Millisecond)/r.reqCount, int(r.ServerProcessingBest/time.Millisecond))

			if !r.t5.IsZero() {
				fmt.Fprintf(&buf, "Content transfer:  %4d ms        %4d ms\n\n",
					int(r.contentTransferSum/time.Millisecond)/r.reqCount, int(r.contentTransferBest/time.Millisecond))
			} else {
				fmt.Fprintf(&buf, "Content transfer:  %4s ms\n", "-")
			}

			fmt.Fprintf(&buf, "Name Lookup:       %4d ms        %4d ms\n",
				int(r.NameLookupSum/time.Millisecond)/r.reqCount, int(r.NameLookupBest/time.Millisecond))
			fmt.Fprintf(&buf, "Connect:           %4d ms        %4d ms\n",
				int(r.ConnectSum/time.Millisecond)/r.reqCount, int(r.ConnectBest/time.Millisecond))
			fmt.Fprintf(&buf, "Pre Transfer:      %4d ms        %4d ms\n",
				int(r.PretransferSum/time.Millisecond)/r.reqCount, int(r.PretransferBest/time.Millisecond))
			fmt.Fprintf(&buf, "Start Transfer:    %4d ms        %4d ms\n",
				int(r.StartTransferSum/time.Millisecond)/r.reqCount, int(r.StartTransferBest/time.Millisecond))

			if !r.t5.IsZero() {
				fmt.Fprintf(&buf, "Total:             %4d ms        %4d ms\n",
					int(r.totalSum/time.Millisecond)/r.reqCount, int(r.totalBest/time.Millisecond))
			} else {
				fmt.Fprintf(&buf, "Total:             %4s ms\n", "-")
			}
			fmt.Fprintf(&buf, "Request count:      %4d\n", r.reqCount)
			io.WriteString(s, buf.String())
			return
		}

		fallthrough
	case 's', 'q':
		d := r.durations()
		list := make([]string, 0, len(d))
		for k, v := range d {
			// Handle when End function is not called
			if (k == "ContentTransfer" || k == "Total") && r.t5.IsZero() {
				list = append(list, fmt.Sprintf("%s: - ms", k))
				continue
			}
			list = append(list, fmt.Sprintf("%s: %d ms", k, v/time.Millisecond))
		}
		io.WriteString(s, strings.Join(list, ", "))
	}

}

func WithClientTrace(ctx context.Context, r *Result) context.Context {
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(i httptrace.DNSStartInfo) {
			r.dnsStart = time.Now()
		},

		DNSDone: func(i httptrace.DNSDoneInfo) {
			r.dnsDone = time.Now()

			r.DNSLookup = r.dnsDone.Sub(r.dnsStart)
			r.NameLookup = r.dnsDone.Sub(r.dnsStart)
		},

		ConnectStart: func(_, _ string) {
			r.tcpStart = time.Now()

			// When connecting to IP (When no DNS lookup)
			if r.dnsStart.IsZero() {
				r.dnsStart = r.tcpStart
				r.dnsDone = r.tcpStart
			}
		},

		ConnectDone: func(network, addr string, err error) {
			r.tcpDone = time.Now()

			r.TCPConnection = r.tcpDone.Sub(r.tcpStart)
			r.Connect = r.tcpDone.Sub(r.dnsStart)
		},

		TLSHandshakeStart: func() {
			r.isTLS = true
			r.tlsStart = time.Now()
		},

		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			r.tlsDone = time.Now()

			r.TLSHandshake = r.tlsDone.Sub(r.tlsStart)
			r.Pretransfer = r.tlsDone.Sub(r.dnsStart)
		},

		GotConn: func(i httptrace.GotConnInfo) {
			// Handle when keep alive is used and connection is reused.
			// DNSStart(Done) and ConnectStart(Done) is skipped
			if i.Reused {
				r.isReused = true
			}
		},

		WroteRequest: func(info httptrace.WroteRequestInfo) {
			r.serverStart = time.Now()

			// When client doesn't use DialContext or using old (before go1.7) `net`
			// pakcage, DNS/TCP/TLS hook is not called.
			if r.dnsStart.IsZero() && r.tcpStart.IsZero() {
				now := r.serverStart

				r.dnsStart = now
				r.dnsDone = now
				r.tcpStart = now
				r.tcpDone = now
			}

			// When connection is re-used, DNS/TCP/TLS hook is not called.
			if r.isReused {
				now := r.serverStart

				r.dnsStart = now
				r.dnsDone = now
				r.tcpStart = now
				r.tcpDone = now
				r.tlsStart = now
				r.tlsDone = now
			}

			if r.isTLS {
				return
			}

			r.TLSHandshake = r.tcpDone.Sub(r.tcpDone)
			r.Pretransfer = r.Connect
		},

		GotFirstResponseByte: func() {
			r.serverDone = time.Now()

			r.ServerProcessing = r.serverDone.Sub(r.serverStart)
			r.StartTransfer = r.serverDone.Sub(r.dnsStart)

			r.transferStart = r.serverDone
		},
	})
}

// End sets the time when reading response is done.
// This must be called after reading response body.
func (r *Result) End(t time.Time) {
	r.trasferDone = t
	r.t5 = t // for Formatter

	// This means result is empty (it does nothing).
	// Skip setting value(contentTransfer and total will be zero).
	if r.dnsStart.IsZero() {
		return
	}

	r.contentTransfer = r.trasferDone.Sub(r.transferStart)
	r.total = r.trasferDone.Sub(r.dnsStart)
	r.reqCount++
	r.sums()
	r.best()
}
