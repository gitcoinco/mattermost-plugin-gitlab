package webhook

import (
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"
)

func (w *webhook) HandleTag(event *gitlab.TagEvent) ([]*HandleWebhook, error) {
	handlers, err := w.handleDMTag(event)
	if err != nil {
		return nil, err
	}
	handlers2, err := w.handleChannelTag(event)
	if err != nil {
		return nil, err
	}
	return cleanWebhookHandlers(append(handlers, handlers2...)), nil
}

func (w *webhook) handleDMTag(event *gitlab.TagEvent) ([]*HandleWebhook, error) {
	senderGitlabUsername := event.UserName
	handlers := []*HandleWebhook{}
	tagNames := strings.Split(event.Ref, "/")
	tagName := tagNames[len(tagNames)-1]

	if mention := w.handleMention(mentionDetails{
		senderUsername:    senderGitlabUsername,
		pathWithNamespace: event.Project.PathWithNamespace,
		IID:               tagName,
		URL:               event.Commits[0].URL,
		body:              event.Message,
	}); mention != nil {
		handlers = append(handlers, mention)
	}

	return handlers, nil
}

func (w *webhook) handleChannelTag(event *gitlab.TagEvent) ([]*HandleWebhook, error) {
	senderGitlabUsername := event.UserName
	repo := event.Project
	tagNames := strings.Split(event.Ref, "/")
	tagName := tagNames[len(tagNames)-1]
	res := []*HandleWebhook{}

	message := fmt.Sprintf("[%s](%s) New tag [%s](%s) by [%s](%s): %s", repo.PathWithNamespace, repo.WebURL, tagName, event.Commits[0].URL, senderGitlabUsername, w.gitlabRetreiver.GetUserURL(senderGitlabUsername), event.Message)

	toChannels := make([]string, 0)
	subs := w.gitlabRetreiver.GetSubscribedChannelsForRepository(repo.PathWithNamespace, repo.Visibility == gitlab.PublicVisibility)
	for _, sub := range subs {
		if !sub.Tag() {
			continue
		}

		toChannels = append(toChannels, sub.ChannelID)
	}

	if len(toChannels) > 0 {
		res = append(res, &HandleWebhook{
			From:       senderGitlabUsername,
			Message:    message,
			ToUsers:    []string{},
			ToChannels: toChannels,
		})
	}

	return res, nil
}
