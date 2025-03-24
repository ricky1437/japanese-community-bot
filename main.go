package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var BotToken string
var session *discordgo.Session

func init() {
	if os.Getenv("GO_ENV") == "dev" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Failed to load environment variables")
		}
	}

	BotToken = os.Getenv("BOT_TOKEN")

	var err error
	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot token: %v", err)
	}

	fmt.Println("Bot is now running.")

	session.AddHandler(onVerifyCommand)
	session.AddHandler(onVerifyButtonClick)
	session.AddHandler(onUnverifyButtonClick)
}

func sendMessage(s *discordgo.Session, channelId string, msg string) {
	_, err := s.ChannelMessageSend(channelId, msg)
	log.Println(">>>" + msg)
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
}

func sendMessageComplex(s *discordgo.Session, channelId string, data *discordgo.MessageSend) {
	_, err := s.ChannelMessageSendComplex(channelId, data)
	if err != nil {
		log.Fatalf("Error sending verification message: %v", err)
		return
	}
}

func onVerifyCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != os.Getenv("ROLE_CHANNEL_ID") {
		return
	}
	u := m.Author
	userPerms, err := s.UserChannelPermissions(u.ID, m.ChannelID)

	var isUserAdmin bool = userPerms&discordgo.PermissionAdministrator != 0

	if err != nil {
		log.Fatalf("Failed to get UserChannelPermission data: %v", err)
		return
	}

	fmt.Printf("%20s %20s(%s) isAdmin:%t > %s\n", m.ChannelID, u.Username, u.ID, isUserAdmin, m.Content)

	{
		verify := discordgo.Button{
			Label:    "追加",
			Style:    discordgo.PrimaryButton,
			CustomID: "verify",
		}

		unverify := discordgo.Button{
			Label:    "解除",
			Style:    discordgo.SecondaryButton,
			CustomID: "unverify",
		}

		actions := discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{verify, unverify},
		}

		content := "スーパーマリオワールドRTA日本Discordサーバーへようこそ！\n走者の方はボタンをクリックしてください。特定チャンネルの閲覧が可能になります。"

		data := &discordgo.MessageSend{
			Components: []discordgo.MessageComponent{actions},
			Content:    content,
		}

		if m.Content != "!verify" {
			return
		}

		if !isUserAdmin {
			log.Printf("[SECURITY WARN]Execution Attempt from unpriviledged user!\n Username: %s ID:%s", u.Username, u.ID)
			sendMessage(s, m.ChannelID, u.Mention()+"このコマンドは管理者のみ使用できます。")
			return
		}
		sendMessageComplex(s, m.ChannelID /* u.Mention() + */, data)
	}

}

func giveRunnerRole(s *discordgo.Session, guildID string, userID string, roleID string) {
	err := s.GuildMemberRoleAdd(guildID, userID, roleID)
	if err != nil {
		log.Fatalf("Error giving a role to an user: %v", err)
	}
}

func onVerifyButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var interactedUserID string
	if i.MessageComponentData().CustomID == "verify" {

		if i.Interaction.Member != nil {
			interactedUserID = i.Interaction.Member.User.ID
		} else if i.Interaction.User != nil {
			interactedUserID = i.Interaction.User.ID
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "走者ロールを付与しました。",
			},
		}

		giveRunnerRole(s, os.Getenv("GUILD_ID"), interactedUserID, os.Getenv("RUNNER_ROLE_ID"))

		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			log.Fatalf("Error responding interaction: %v", err)
		}
	}
}

func removeRunnerRole(s *discordgo.Session, guildID string, userID string, roleID string) (r bool) {
	m, err := s.GuildMember(guildID, userID)
	if err != nil {
		log.Fatalf("Error resoponding interaction: %v", err)
		return
	}

	var isRoleMatched bool

	for _, v := range m.Roles {
		if v == roleID {
			isRoleMatched = true

			fmt.Printf("User role matched! %v", v)
			err = s.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				log.Fatalf("Error removing role: %v", err)
			}
			return isRoleMatched
		}

		fmt.Println("User role not matched.")
		isRoleMatched = false
		continue
	}

	return isRoleMatched
}

func onUnverifyButtonClick(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var interactedUserID string
	if i.MessageComponentData().CustomID == "unverify" {
		if i.Interaction.Member != nil {
			interactedUserID = i.Interaction.Member.User.ID
		} else if i.Interaction.User != nil {
			interactedUserID = i.Interaction.User.ID
		}

		r := removeRunnerRole(s, os.Getenv("GUILD_ID"), interactedUserID, os.Getenv("RUNNER_ROLE_ID"))

		var content string

		if r == false {
			content = "走者ロールはすでに解除されています。"
		} else if r == true {
			content = "走者ロールを解除しました。"
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: content,
			},
		}

		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			log.Fatalf("Error responding interaction: %v", err)
		}
	}
}

func main() {
	err := session.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	fmt.Println("Connection established.")
	select {}
}
