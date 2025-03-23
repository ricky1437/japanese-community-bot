package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Env struct {
	ClientId  string
	Token     string
	GuildId   string
	ChannelId string
	RoleId    string
}

var env Env

func loadEnv(env *Env) {
	err := godotenv.Load()

	env.ClientId = os.Getenv("CLIENT_ID")
	env.Token = os.Getenv("TOKEN")
	env.GuildId = os.Getenv("GUILD_ID")
	env.ChannelId = os.Getenv("CHANNEL_ID")
	env.RoleId = os.Getenv("ROLE_ID")

	if err != nil {
		fmt.Printf("Could not load .env: %v", err)
	}

	fmt.Println("Sucessfully .env loaded.")
}

func sendMessage(s *discordgo.Session, channelId string, msg string) {
	_, err := s.ChannelMessageSend(channelId, msg)
	log.Println(">>>" + msg)
	if err != nil {
		log.Println("Error sending message: ", err)
	}
}

func sendMessageComplex(s *discordgo.Session, channelId string, data *discordgo.MessageSend) {
	_, err := s.ChannelMessageSendComplex(channelId, data)
	if err != nil {
		log.Println("Error sending verification message: ", err)
		return
	}
}

func onVerifyCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.ChannelID != env.ChannelId {
		return
	}
	u := m.Author
	userPerms, err := s.UserChannelPermissions(u.ID, m.ChannelID)

	var isUserAdmin bool = userPerms&discordgo.PermissionAdministrator != 0

	if err != nil {
		fmt.Printf("Failed to get UserChannelPermission data: %v", err)
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
			fmt.Printf("[SECURITY WARN]Execution Attempt from unpriviledged user!\n Username: %s ID:%20s", u.Username, u.ID)
			sendMessage(s, m.ChannelID, u.Mention()+"このコマンドは管理者のみ使用できます。")
			return
		}
		sendMessageComplex(s, m.ChannelID /* u.Mention() + */, data)
	}

}

func giveRunnerRole(s *discordgo.Session, guildID string, userID string, roleID string) {
	err := s.GuildMemberRoleAdd(guildID, userID, roleID)
	if err != nil {
		log.Println("Error giving a role to an user.")
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

		giveRunnerRole(s, env.GuildId, interactedUserID, env.RoleId)

		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			log.Println("Error responding interaction: ", err)
		}
	}
}

func removeRunnerRole(s *discordgo.Session, guildID string, userID string, roleID string) (r bool) {
	m, err := s.GuildMember(guildID, userID)
	if err != nil {
		log.Println("Error resoponding interaction: ", err)
		return
	}

	var isRoleMatched bool

	for _, v := range m.Roles {
		if v == roleID {
			isRoleMatched = true

			fmt.Printf("User role matched! %s", v)
			err = s.GuildMemberRoleRemove(guildID, userID, roleID)
			if err != nil {
				fmt.Printf("Error removing role: %s", err)
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

		r := removeRunnerRole(s, env.GuildId, interactedUserID, env.RoleId)

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
			log.Println("Error responding interaction: ", err)
		}
	}
}

func main() {
	loadEnv(&env)

	session, err := discordgo.New("Bot " + env.Token)
	if err != nil {
		fmt.Println("Failed to log in,", err)
	}

	session.AddHandler(onVerifyCommand)
	session.AddHandler(onVerifyButtonClick)
	session.AddHandler(onUnverifyButtonClick)

	err = session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	select {}
}
