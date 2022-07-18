package main

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"strings"
	"tg_bot_mafia_stat/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	Conn  *sqlx.DB
	Bot   *tgbotapi.BotAPI
	Table models.Table
)

func main() {
	var err error
	token := os.Getenv("TOKEN")
	Bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	login := os.Getenv("mafia_db_login")
	password := os.Getenv("mafia_db_password")
	connStr := "postgres://" + login + ":" + password + "@localhost:5432/mafia"

	Conn, err = sqlx.Open("postgres", connStr)
	if err != nil {
		log.Fatalln("Can not connect to mafia db:", err)
	}
	defer Conn.Close()

	Bot.Debug = true

	log.Printf("Authorized on account %s", Bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := Bot.GetUpdatesChan(u)

	adminID, err := strconv.Atoi(os.Getenv("ADMIN_TG_ID"))

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.From.ID == int64(adminID) {
				adminCommands(&update)
				continue
			}

			playerCommands(&update)
		}
	}
}

func playerCommands(update *tgbotapi.Update) {
	switch update.Message.Command() {
	case "sit_1", "sit_2", "sit_3", "sit_4", "sit_5", "sit_6", "sit_7", "sit_8", "sit_9", "sit_10":
		if !Table.Open {
			sendMsg(update, "Стол закрыт, выбор позиции недоступен")
			return
		}
		words := strings.Split(update.Message.Text, "_")
		if len(words) != 2 {
			sendMsg(update, "Введите номер вашей позиции за столом:\n\n /sit <1-10>")
			return
		}
		pos, err := strconv.Atoi(words[1])
		if err != nil || pos > 10 || pos < 1 {
			if err == nil {
				err = errors.New(words[1])
			}
			sendMsg(update, "Неправильное число: "+err.Error()+"\n"+
				"Введите номер вашей позиции за столом:\n\n /sit <1-10>")
			return
		}
		updatePosTable(words[1], update.Message.From.ID)
		sendMsg(update, "Вы заняли слот № "+words[1])
	case "slack":
		words := strings.Split(update.Message.Text, " ")
		if len(words) != 2 {
			sendMsg(update, "Введите ваш ник в слаке:\n\n/slack <NICKNAME>")
			return
		}
		err := updateSlackNickname(update, words[1])
		if err != nil {
			sendMsg(update, "Никнейм не обновлен: "+err.Error())
			return
		}
		sendMsg(update, "Никнейм был немедленно обновлен")
	}
}

func updateSlackNickname(update *tgbotapi.Update, slackNickname string) error {
	_, err := Conn.Exec("select update_slack_nickname($1, $2);", update.Message.From.ID, slackNickname)
	return err
}

func updatePosTable(pos string, telegramID int64) {
	query := fmt.Sprintf("update \"table\" set pos_" + pos + "=$1 where id = 1;")
	_, err := Conn.Exec(query, telegramID)
	if err != nil {
		log.Println("update pos "+pos+" table:", err)
	} else {
		log.Println("pos " + pos + " table was updated")
	}
}

func sendMsg(update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)

	Bot.Send(msg)
}

func adminCommands(update *tgbotapi.Update) {
	var err error
	switch update.Message.Command() {
	case "open_table":
		if Table.GameStarted {
			sendMsg(update, "Игра уже началась")
			return
		}
		sendMsg(update, "Стол открыт")
		Table.Open = true
	case "drop_table":
		sendMsg(update, "Стол обновлен")
		dropTable()
	case "check_nicknames":
		sendMsg(update, "Проверка ников")
		var destTable []models.Table
		fmt.Printf("%v\n", destTable)
		err = Conn.Select(&destTable, "select * from \"table\";")
		if err != nil {
			log.Println("check nicknames:", err)
		}
		Table = destTable[0]
		sendMsg(update, fmt.Sprintf("%#v\n", Table))
	case "close_table":
		sendMsg(update, "Стол закрыт")
		Table.Open = false
	case "set_mafia":
		if Table.Open {
			sendMsg(update, "Необходимо закрыть стол")
			return
		}
		words := strings.Split(update.Message.Text, " ")
		if len(words) != 4 {
			sendMsg(update, "Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		mafiaBoss, err := strconv.Atoi(words[1])
		if err != nil || mafiaBoss > 10 || mafiaBoss < 1 {
			sendMsg(update, "Неправильное число: "+words[1]+"\n"+
				"Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		mafia1, err := strconv.Atoi(words[2])
		if err != nil || mafia1 > 10 || mafia1 < 1 {
			sendMsg(update, "Неправильное число: "+words[2]+"\n"+
				"Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		mafia2, err := strconv.Atoi(words[3])
		if err != nil || mafia2 > 10 || mafia2 < 1 {
			sendMsg(update, "Неправильное число: "+words[3]+"\n"+
				"Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		Table.MafiaBoss = mafiaBoss
		Table.Mafia1 = mafia1
		Table.Mafia2 = mafia2
		_, err = Conn.Exec("update \"table\" set mafia_boss=$1, mafia_1=$2, mafia_2=$3 where id = 1;", Table.MafiaBoss, Table.Mafia1, Table.Mafia2)
		if err != nil {
			sendMsg(update, "err: "+err.Error())
			return
		}
		Table.MafiaStored = true
		sendMsg(update, "Позиции мафии записаны")
	case "set_sherif":
		if Table.Open {
			sendMsg(update, "Необходимо закрыть стол")
			return
		}

		words := strings.Split(update.Message.Text, " ")
		if len(words) != 4 {
			sendMsg(update, "Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		sherif, err := strconv.Atoi(words[1])
		if err != nil || sherif > 10 || sherif < 1 {
			sendMsg(update, "Неправильное число: "+words[1]+"\n"+
				"Введите 3 числа от 1 до 10 с позициями игроков в таком порядке: Дон, Мафия, Мафия")
			return
		}
		Table.Sherif = sherif
		_, err = Conn.Exec("update \"table\" set sherif=$1 where id = 1;", Table.Sherif)
		if err != nil {
			sendMsg(update, "err: "+err.Error())
			return
		}
		Table.SherifStored = true
		sendMsg(update, "Позиция шерифа записана")
	case "start_game":
		if Table.Open {
			sendMsg(update, "Необходимо закрыть стол")
			return
		}
		if !Table.SherifStored {
			sendMsg(update, "Необходимо сохранить позицию шерифа")
			return
		}
		if !Table.MafiaStored {
			sendMsg(update, "Необходимо сохранить позицию мафии")
			return
		}
		err = startGame()
		if err != nil {
			sendMsg(update, "Не получилось запустить игру: "+err.Error())
			return
		}
		Table.GameStarted = true
		sendMsg(update, "Игра началась")
	case "stop_game":
		Table.GameStarted = false
		sendMsg(update, "Игра остановлена")
	case "update_stats":

	}
}

func startGame() error {
	_, err := Conn.Exec("select start_game();")
	return err
}

func dropTable() {
	Table = models.Table{
		Pos1:         -1,
		Pos2:         -1,
		Pos3:         -1,
		Pos5:         -1,
		Pos6:         -1,
		Pos7:         -1,
		Pos8:         -1,
		Pos9:         -1,
		Pos10:        -1,
		Mafia1:       -1,
		Mafia2:       -1,
		MafiaBoss:    -1,
		Sherif:       -1,
		MafiaStored:  false,
		SherifStored: false,
		Open:         false,
		GameStarted:  false,
	}

	_, err := Conn.Exec("update \"table\" set pos_1=$1, pos_2=$2, pos_3=$3, pos_4=$4, pos_5=$5, pos_6=$6, pos_7=$7,"+
		" pos_8=$8, pos_9=$9, pos_10=$10, mafia_1=$11, mafia_2=$12, mafia_boss=$13, sherif=$14 where id = 1;",
		Table.Pos1,
		Table.Pos2,
		Table.Pos3,
		Table.Pos4,
		Table.Pos5,
		Table.Pos6,
		Table.Pos7,
		Table.Pos8,
		Table.Pos9,
		Table.Pos10,
		Table.Mafia1,
		Table.Mafia2,
		Table.MafiaBoss,
		Table.Sherif,
	)

	if err != nil {
		log.Println("table was not dropped:", err)
	} else {
		log.Println("table was dropped")
	}
}
