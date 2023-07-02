package database

func UsedStories(wallet string) (int, error) {
	var number int

	row := db.QueryRow("SELECT count(stories.id) FROM stories, users WHERE stories.user_id = users.id and users.wallet = $1 and stories.created >= NOW() - INTERVAL '24 HOURS';", wallet)
	err := row.Scan(&number)

	return number, err
}

func BonusStories(wallet string) (int, error) {
	var number int

	row := db.QueryRow("SELECT bonus - users.used_bonus FROM users WHERE wallet = $1;", wallet)
	err := row.Scan(&number)

	return number, err
}

func UseBonus(wallet string) error {
	_, err := db.Exec("UPDATE users SET used_bonus = used_bonus + 1 WHERE wallet = $1;", wallet)

	return err
}

func NewStory(wallet string, storyId string) error {
	_, err := db.Exec("INSERT INTO stories(id, user_id) VALUES ($1, (SELECT id from users where wallet = $2));", storyId, wallet)

	return err
}

func SetStep(storyId string, step int) error {
	_, err := db.Exec("UPDATE stories SET step = $2, image = false WHERE id = $1;", storyId, step)

	return err
}

func ImageGenerated(storyId string) (bool, error) {
	var generated bool

	row := db.QueryRow("SELECT get_image($1);", storyId)
	err := row.Scan(&generated)

	return generated, err
}

func GetStep(storyId string) (int, error) {
	var step int

	row := db.QueryRow("SELECT step FROM stories WHERE id = $1;", storyId)
	err := row.Scan(&step)

	return step, err
}
