package ffmpeg

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/disgoorg/disgo/voice"
	"github.com/jonas747/ogg"
)

const (
	Exec       = "ffmpeg"
	Channels   = 1
	SampleRate = 48000
	BufferSize = 65307
)

var _ voice.OpusFrameProvider = (*AudioProvider)(nil)

func New(ctx context.Context, r io.Reader) (*AudioProvider, error) {
	// Create a pipe for ffmpeg output
	pr, pw := io.Pipe()

	// Start ffmpeg process to transcode input to OGG/Opus
	go func() {
		cmd := exec.CommandContext(ctx, Exec,
			"-i", "pipe:0",
			"-f", "ogg",
			"-c:a", "libopus",
			"-ar", fmt.Sprintf("%d", SampleRate),
			"-ac", fmt.Sprintf("%d", Channels),
			"-b:a", "128k",
			"pipe:1",
		)
		cmd.Stdin = r
		cmd.Stdout = pw
		// cmd.Stderr = os.Stderr // For debugging

		_ = cmd.Run()
		pw.Close()
	}()

	done, doneFunc := context.WithCancel(context.Background())
	return &AudioProvider{
		source:   pr,
		d:        ogg.NewPacketDecoder(ogg.NewDecoder(bufio.NewReaderSize(pr, BufferSize))),
		done:     done,
		doneFunc: doneFunc,
	}, nil
}

type AudioProvider struct {
	source   io.Reader
	d        *ogg.PacketDecoder
	done     context.Context
	doneFunc context.CancelFunc
}

func (p *AudioProvider) ProvideOpusFrame() ([]byte, error) {
	data, _, err := p.d.Decode()
	if err != nil {
		// Only log unexpected errors, not EOF or closed pipe
		if errors.Is(err, io.EOF) || errors.Is(err, os.ErrClosed) || err.Error() == "io: read/write on closed pipe" {
			p.doneFunc()
			return nil, io.EOF
		}
		return nil, fmt.Errorf("error decoding ogg packet: %w", err)
	}
	return data, nil
}

func (p *AudioProvider) Close() {
	if c, ok := p.source.(io.Closer); ok {
		_ = c.Close()
	}
	p.doneFunc()
}

func (p *AudioProvider) Wait() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-p.done.Done()
	}()
	wg.Wait()
	return nil
}
