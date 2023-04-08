package bot

import (
	tb "gopkg.in/telebot.v3"
)

func (bc *BotController) configHandleAccountDelete(m *tb.Message, params ...string) {
	bc.Logf(INFO, m, "User issued account deletion command")
	if len(params) == 1 && params[0] == "yes" {
		bc.Logf(INFO, m, "Will delete all user data upon user request")

		bc.DeleteUserData(m)

		bc.Bot.SendSilent(bc, Recipient(m), "I'm sad to see you go. Hopefully one day, you will come back.\n\nI have deleted all of your data stored in the bot. You can simply start over by sending me a message again. Goodbye.")
		bc.Bot.SendSilent(bc, Recipient(m), "============")
		return
	}
	bc.Logf(INFO, m, "Reset command failed 'yes' verification. Aborting.")
	bc.Bot.SendSilent(bc, Recipient(m), "Reset has been aborted.\n\nYou tried to permanently delete your account. Please make sure to confirm this action by adding 'yes' to the end of your command. Please check /config for usage.")
}

func (bc *BotController) DeleteUserData(m *tb.Message) {
	errors := errors{operation: "user deletion", bc: bc, m: m}
	errors.handle1(bc.Repo.DeleteAllCacheEntries(m))

	errors.handle1(bc.Repo.UserSetNotificationSetting(m, -1, -1))

	errors.handle1(bc.Repo.DeleteTransactions(m))
	errors.handle1(bc.Repo.DeleteTemplates(m))

	errors.handle1(bc.Repo.DeleteAllUserSettings(m.Chat.ID))

	bc.State.Clear(m)
	errors.handle1(bc.Repo.DeleteUser(m))
}
