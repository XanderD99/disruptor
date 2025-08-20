package lavalink

import (
	"context"
	"log/slog"
	"time"

	"github.com/disgoorg/disgolink/v3/disgolink"
	disgolavalink "github.com/disgoorg/disgolink/v3/lavalink"
	"golang.org/x/sync/errgroup"

	"github.com/XanderD99/disruptor/internal/disruptor"
	"github.com/XanderD99/disruptor/internal/metrics"
)

type Node struct {
	// 🏷️ Name of the Lavalink node (must be unique)
	Name string `env:"NAME" default:"disruptor"`
	// 🌐 Lavalink server address (e.g., localhost:2333)
	Address string `env:"ADDRESS" default:"localhost:2333"`
	// 🔑 Lavalink server password
	Password string `env:"PASSWORD"`
	// 🔒 Use secure connection (wss)
	Secure bool `env:"SECURE" default:"false"`
}

type Lavalink interface {
	disgolink.Client
	Start(ctx context.Context) error
}

type lavalink struct {
	disgolink.Client
	nodes []Node
}

func New(nodes []Node, session *disruptor.Session, logger *slog.Logger) Lavalink {
	client := disgolink.New(
		session.ApplicationID(),
		disgolink.WithLogger(logger),
		disgolink.WithListenerFunc(onTrackEnd(session, logger)),
		disgolink.WithListenerFunc(onTrackStart(logger)),
	)

	return &lavalink{
		Client: client,
		nodes:  nodes,
	}
}

func onTrackStart(logger *slog.Logger) func(disgolink.Player, disgolavalink.TrackStartEvent) {
	return func(_ disgolink.Player, event disgolavalink.TrackStartEvent) {
		guildID := event.GuildID()
		logger = logger.With(
			slog.String("guild.id", guildID.String()),
			slog.Group("track",
				slog.String("identifier", event.Track.Info.Identifier),
				slog.Duration("duration", time.Duration(event.Track.Info.Length)),
			),
		)

		logger.Info("track started")

		// Record audio track event
		audioMetrics := metrics.NewAudioMetrics()
		audioMetrics.RecordTrackEvent("start", guildID)
	}
}

func onTrackEnd(session *disruptor.Session, logger *slog.Logger) func(disgolink.Player, disgolavalink.TrackEndEvent) {
	return func(_ disgolink.Player, event disgolavalink.TrackEndEvent) {
		guildID := event.GuildID()
		logger = logger.With(
			slog.String("guild.id", guildID.String()),
			slog.Group("track",
				slog.String("identifier", event.Track.Info.Identifier),
				slog.Duration("duration", time.Duration(event.Track.Info.Length)),
			),
		)

		logger.Info("track ended")

		// Record audio track event and processing duration
		audioMetrics := metrics.NewAudioMetrics()
		audioMetrics.RecordTrackEvent("end", guildID)

		// Calculate sleep duration as 2% of track duration, max 500ms
		trackDuration := time.Duration(event.Track.Info.Length) * time.Millisecond
		// 2% of track duration
		// Cap at 500ms maximum
		// Minimum of 250ms to ensure some delay
		sleepDuration := max(min(time.Duration(float64(trackDuration)*0.02), 500*time.Millisecond), 250*time.Millisecond)

		logger.Debug("waiting before leaving voice channel", slog.Duration("sleep", sleepDuration))

		// Record audio processing duration for the sleep/cleanup operation
		timer := audioMetrics.NewAudioProcessingTimer("cleanup", guildID)
		time.Sleep(sleepDuration)

		if err := session.UpdateVoiceState(context.Background(), guildID, nil); err != nil {
			logger.Error("failed to update voice state after track end", slog.Any("error", err))
			audioMetrics.RecordVoiceStateUpdate(guildID, false)
		} else {
			audioMetrics.RecordVoiceStateUpdate(guildID, true)
		}

		timer.Finish()
	}
}

func (l *lavalink) Start(ctx context.Context) error {
	var eg errgroup.Group
	eg.SetLimit(5)

	// Add Lavalink nodes to the client
	for _, node := range l.nodes {
		eg.Go(func() error {
			_, err := l.AddNode(ctx, disgolink.NodeConfig{
				Name:     node.Name,     // a unique node name
				Address:  node.Address,  // e.g., "localhost:2333"
				Password: node.Password, // Lavalink server password
				Secure:   node.Secure,   // ws or wss
			})

			return err
		})
	}

	return eg.Wait()
}
