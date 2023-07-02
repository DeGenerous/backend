package database

func UsedStories(wallet string) (int, error) {
	var number int

	row := db.QueryRow("SELECT count(stories.id) FROM stories, users WHERE stories.user_id = users.id and users.wallet = $1 and stories.created >= NOW() - INTERVAL '24 HOURS' and stories.bonus = false;", wallet)
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

func GeneratePrompt(category string) (int, string, error) {
	var id int
	var prompt string

	row := db.QueryRow("SELECT id, prompt FROM get_prompt($1) AS (id bigint, prompt text);", category)
	err := row.Scan(&id, &prompt)

	return id, prompt, err
}

func GetPrompt(id int) (string, error) {
	var prompt string

	row := db.QueryRow("SELECT prompt from prompts where id = $1;", id)
	err := row.Scan(&prompt)

	return prompt, err
}

func NewStory(wallet string, storyId string, bonus bool) error {
	_, err := db.Exec("INSERT INTO stories(id, user_id, bonus) VALUES ($1, (SELECT id from users where wallet = $2), $3);", storyId, wallet, bonus)

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
