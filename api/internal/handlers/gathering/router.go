package gathering

import (
	"github.com/alexmarian/apc/api/internal/handlers"
	gatheringHandlers "github.com/alexmarian/apc/api/internal/handlers/gathering/handlers"
)

// GatheringRouter provides all gathering-related HTTP handlers
type GatheringRouter struct {
	Gathering    *gatheringHandlers.GatheringHandler
	VotingMatter *gatheringHandlers.VotingMatterHandler
	Participant  *gatheringHandlers.ParticipantHandler
	Ballot       *gatheringHandlers.BallotHandler
	Results      *gatheringHandlers.ResultsHandler
	Export       *gatheringHandlers.ExportHandler
	Notification *gatheringHandlers.NotificationHandler
}

// NewGatheringRouter creates and initializes all gathering handlers
func NewGatheringRouter(cfg *handlers.ApiConfig) *GatheringRouter {
	// Create gathering handler first as others depend on it
	gatheringHandler := gatheringHandlers.NewGatheringHandler(cfg)

	return &GatheringRouter{
		Gathering:    gatheringHandler,
		VotingMatter: gatheringHandlers.NewVotingMatterHandler(cfg, gatheringHandler),
		Participant:  gatheringHandlers.NewParticipantHandler(cfg, gatheringHandler),
		Ballot:       gatheringHandlers.NewBallotHandler(cfg, gatheringHandler),
		Results:      gatheringHandlers.NewResultsHandler(cfg),
		Export:       gatheringHandlers.NewExportHandler(cfg),
		Notification: gatheringHandlers.NewNotificationHandler(cfg),
	}
}
