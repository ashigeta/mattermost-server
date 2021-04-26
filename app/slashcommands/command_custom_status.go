// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"strings"

	goi18n "github.com/mattermost/go-i18n/i18n"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

type CustomStatusProvider struct {
}

const (
	CmdCustomStatus      = app.CmdCustomStatusTrigger
	CmdCustomStatusClear = "clear"
        CmdCustomStatusPrev  = "prev"
        CmdCustomStatusAway  = "away"
        CmdCustomStatusHome  = "home"
        CmdCustomStatusOffice = "office"
        CmdCustomStatusMeeting = "meeting"

	DefaultCustomStatusEmoji = "speech_balloon"
)

func init() {
	app.RegisterCommandProvider(&CustomStatusProvider{})
}

func (*CustomStatusProvider) GetTrigger() string {
	return CmdCustomStatus
}

func (*CustomStatusProvider) GetCommand(a *app.App, T goi18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CmdCustomStatus,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_custom_status.desc"),
		AutoCompleteHint: T("api.command_custom_status.hint"),
                AutocompleteData: GetCommandAutoCompleteData(T),
		DisplayName:      T("api.command_custom_status.name"),
	}
}

func GetCommandAutoCompleteData(T goi18n.TranslateFunc) *model.AutocompleteData {
        command := model.NewAutocompleteData(CmdCustomStatus, "", T("api.command_custom_status.desc"))
        command.AddStaticListArgument("", true, []model.AutocompleteListItem{
                {
                        Item: CmdCustomStatusClear,
                        HelpText: "Clear your custom status",
                }, {
                        Item: CmdCustomStatusPrev,
                        HelpText: "Set previous custom status",
                }, {
                        Item: CmdCustomStatusAway,
                        Hint: "[HH:MM] [message]",
                        HelpText: "Set your custom status to 'Away'",
                }, {
                        Item: CmdCustomStatusHome,
                        Hint: "[message]",
                        HelpText: "Set your custom status to 'Working from home'",
                }, {
                        Item: CmdCustomStatusOffice,
                        Hint: "[message]",
                        HelpText: "Set your custom status to 'At the office'",
                }, {
                        Item: CmdCustomStatusMeeting,
                        Hint: "[HH:MM] [message]",
                        HelpText: "Set your custom status to 'In a meeting'",
                }, {
                        Hint: "[:emoji_name:] [message]",
                        HelpText: "Set custom status",
                },
        })
        return command
}

func (*CustomStatusProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	if !a.Config().FeatureFlags.CustomUserStatuses || !*a.Config().TeamSettings.EnableCustomUserStatuses {
		return nil
	}

	if strings.HasPrefix(message, CmdCustomStatusClear) {
		if err := a.RemoveCustomStatus(args.UserId); err != nil {
			mlog.Error(err.Error())
			return &model.CommandResponse{Text: args.T("api.command_custom_status.clear.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
		}

		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         args.T("api.command_custom_status.clear.success"),
		}
	}
        if strings.HasPrefix(message, CmdCustomStatusPrev) {
                if err := a.SetPrevRecentCustomStatus(args.UserId); err != nil {
			mlog.Error(err.Error())
			return &model.CommandResponse{Text: args.T("api.command_custom_status.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
                }
                cs, err := a.GetCustomStatus(args.UserId)
                if err != nil {
			mlog.Error(err.Error())
			return &model.CommandResponse{Text: args.T("api.command_custom_status.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
                }
                if cs == nil {
                        return &model.CommandResponse{
                                ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
                                Text:         args.T("api.command_custom_status.clear.success"),
                        }
                }
                return &model.CommandResponse{
                        ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
                        Text: args.T("api.command_custom_status.success", map[string]interface{}{
                                "EmojiName":     ":" + cs.Emoji + ":",
                                "StatusMessage": cs.Text,
                        }),
                }
        }

	customStatus := &model.CustomStatus{
		Emoji: DefaultCustomStatusEmoji,
		Text:  message,
	}

        if strings.HasPrefix(message, CmdCustomStatusAway) {
                customStatus.Emoji = "car"
                customStatus.Text  = "Away"
                msg := strings.TrimPrefix(message, CmdCustomStatusAway)
                msg  = strings.TrimSpace(msg)
                if msg != "" {
                        customStatus.Text += " (" + msg + ")"
                }
        } else if strings.HasPrefix(message, CmdCustomStatusHome) {
                customStatus.Emoji = "house"
                customStatus.Text  = "Working from home"
                msg := strings.TrimPrefix(message, CmdCustomStatusHome)
                msg  = strings.TrimSpace(msg)
                if msg != "" {
                        customStatus.Text += " (" + msg + ")"
                }
        } else if strings.HasPrefix(message, CmdCustomStatusOffice) {
                customStatus.Emoji = "office"
                customStatus.Text  = "At the office"
                msg := strings.TrimPrefix(message, CmdCustomStatusOffice)
                msg  = strings.TrimSpace(msg)
                if msg != "" {
                        customStatus.Text += " (" + msg + ")"
                }
        } else if strings.HasPrefix(message, CmdCustomStatusMeeting) {
                customStatus.Emoji = "telephone"
                customStatus.Text  = "In a meeting"
                msg := strings.TrimPrefix(message, CmdCustomStatusMeeting)
                msg  = strings.TrimSpace(msg)
                if msg != "" {
                        customStatus.Text += " (" + msg + ")"
                }
        } else {
                firstEmojiLocations := model.ALL_EMOJI_PATTERN.FindIndex([]byte(message))
                if len(firstEmojiLocations) > 0 && firstEmojiLocations[0] == 0 {
                        // emoji found at starting index
                        customStatus.Emoji = message[firstEmojiLocations[0]+1 : firstEmojiLocations[1]-1]
                        customStatus.Text = strings.TrimSpace(message[firstEmojiLocations[1]:])
                }
        }

	customStatus.TrimMessage()
	if err := a.SetCustomStatus(args.UserId, customStatus); err != nil {
		mlog.Error(err.Error())
		return &model.CommandResponse{Text: args.T("api.command_custom_status.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: args.T("api.command_custom_status.success", map[string]interface{}{
			"EmojiName":     ":" + customStatus.Emoji + ":",
			"StatusMessage": customStatus.Text,
		}),
	}
}
