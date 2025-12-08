package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// I18nService handles internationalization for voting results
type I18nService struct {
	translations map[string]map[string]string
	defaultLang  string
}

// Translation keys
const (
	KeyGatheringTitle           = "gathering.title"
	KeyGatheringType            = "gathering.type"
	KeyGatheringTypeInitial     = "gathering.type.initial"
	KeyGatheringTypeRepeated    = "gathering.type.repeated"
	KeyGatheringTypeRemote      = "gathering.type.remote"
	KeyVotingMode               = "voting.mode"
	KeyVotingModeByWeight       = "voting.mode.by_weight"
	KeyVotingModeByUnit         = "voting.mode.by_unit"
	KeyQuorumInfo               = "quorum.info"
	KeyQuorumMet                = "quorum.met"
	KeyQuorumNotMet             = "quorum.not_met"
	KeyQuorumRequired           = "quorum.required"
	KeyQuorumAchieved           = "quorum.achieved"
	KeyQuorumThreshold          = "quorum.threshold"
	KeyResultsPassed            = "results.passed"
	KeyResultsFailed            = "results.failed"
	KeyStatisticsQualifiedUnits = "statistics.qualified_units"
	KeyStatisticsParticipating  = "statistics.participating_units"
	KeyStatisticsVoted          = "statistics.voted_units"
	KeyStatisticsWeight         = "statistics.weight"
	KeyStatisticsArea           = "statistics.area"
	KeyVoteYes                  = "vote.yes"
	KeyVoteNo                   = "vote.no"
	KeyVoteAbstain              = "vote.abstain"
	KeyMatterTitle              = "matter.title"
	KeyMatterType               = "matter.type"
	KeyMatterResult             = "matter.result"
	KeyGeneratedAt              = "generated_at"
)

// NewI18nService creates a new I18nService
func NewI18nService(localesDir string) (*I18nService, error) {
	service := &I18nService{
		translations: make(map[string]map[string]string),
		defaultLang:  "en",
	}

	// Load English translations (default)
	if err := service.loadTranslations("en", localesDir); err != nil {
		// If loading from file fails, use built-in English translations
		service.translations["en"] = getBuiltInEnglishTranslations()
	}

	return service, nil
}

// Translate returns the translation for a key in the specified language
func (s *I18nService) Translate(key string, lang string) string {
	// Try requested language
	if translations, ok := s.translations[lang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}

	// Fallback to default language
	if translations, ok := s.translations[s.defaultLang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}

	// Fallback to key itself
	return key
}

// loadTranslations loads translations from a JSON file
func (s *I18nService) loadTranslations(lang string, localesDir string) error {
	filePath := filepath.Join(localesDir, fmt.Sprintf("%s.json", lang))

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read translation file: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse translation file: %w", err)
	}

	s.translations[lang] = translations
	return nil
}

// getBuiltInEnglishTranslations returns built-in English translations
func getBuiltInEnglishTranslations() map[string]string {
	return map[string]string{
		KeyGatheringTitle:           "Gathering",
		KeyGatheringType:            "Type",
		KeyGatheringTypeInitial:     "Initial Gathering",
		KeyGatheringTypeRepeated:    "Repeated Gathering",
		KeyGatheringTypeRemote:      "Remote Gathering",
		KeyVotingMode:               "Voting Mode",
		KeyVotingModeByWeight:       "By Weight",
		KeyVotingModeByUnit:         "By Unit",
		KeyQuorumInfo:               "Quorum Information",
		KeyQuorumMet:                "Quorum Met",
		KeyQuorumNotMet:             "Quorum Not Met",
		KeyQuorumRequired:           "Required",
		KeyQuorumAchieved:           "Achieved",
		KeyQuorumThreshold:          "Threshold",
		KeyResultsPassed:            "Passed",
		KeyResultsFailed:            "Failed",
		KeyStatisticsQualifiedUnits: "Qualified Units",
		KeyStatisticsParticipating:  "Participating Units",
		KeyStatisticsVoted:          "Voted Units",
		KeyStatisticsWeight:         "Weight",
		KeyStatisticsArea:           "Area",
		KeyVoteYes:                  "Yes",
		KeyVoteNo:                   "No",
		KeyVoteAbstain:              "Abstain",
		KeyMatterTitle:              "Matter",
		KeyMatterType:               "Type",
		KeyMatterResult:             "Result",
		KeyGeneratedAt:              "Generated At",
	}
}

// FormatGatheringType returns the translated gathering type
func (s *I18nService) FormatGatheringType(gatheringType string, lang string) string {
	switch gatheringType {
	case "initial":
		return s.Translate(KeyGatheringTypeInitial, lang)
	case "repeated":
		return s.Translate(KeyGatheringTypeRepeated, lang)
	case "remote":
		return s.Translate(KeyGatheringTypeRemote, lang)
	default:
		return gatheringType
	}
}

// FormatVotingMode returns the translated voting mode
func (s *I18nService) FormatVotingMode(votingMode string, lang string) string {
	switch votingMode {
	case "by_weight":
		return s.Translate(KeyVotingModeByWeight, lang)
	case "by_unit":
		return s.Translate(KeyVotingModeByUnit, lang)
	default:
		return votingMode
	}
}
