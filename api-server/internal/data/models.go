package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		User:  User{},
		Token: Token{},
	} // типа рефреш
}

type Models struct {
	User  User
	Token Token
}

type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Token     Token  `json:"token"`
}

func (u *User) GetAll() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, first_name, last_name, user_active, password, created_at, updated_at,
       CASE 
           WHEN(SELECT count(id) FROM tokens t WHERE user_id = users.id AND t.expiry > NOW()) > 0 THEN 1
		ELSE 0
		END AS has_token 
       FROM users ORDER BY last_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Token.ID,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}
	return users, nil
}

func (u *User) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, first_name, last_name, password, created_at, updated_at, user_active FROM users WHERE email = $1`

	var user User

	row := db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Active,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *User) GetById(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, first_name, last_name, password, created_at, updated_at, user_active FROM users WHERE id = $1`

	var user User

	row := db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Active,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *User) Update() error { // Update метод который по ID переписывает текущие значения
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users SET
		email = $1,
		first_name = $2,
		last_name = $3,
		user_active = $4,
		updated_at = $5
		WHERE id = $6`

	_, err := db.ExecContext(ctx, stmt,
		u.Email,
		u.FirstName,
		u.LastName,
		u.Active,
		time.Now(),
		u.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := db.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) DeleteById(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `INSERT INTO users (email, first_name, last_name, password, user_active, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err = db.QueryRowContext(ctx, stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		hashedPassword,
		user.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (u *User) ResetPassword(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET password = $1 where id = $2`

	_, err = db.ExecContext(ctx, stmt, hashedPassword, &u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

type Token struct {
	ID        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	TokenHash []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Expiry    time.Time `json:"expiry"`
}

func (t *Token) GetByToken(plainText string) (*Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, user_id, email, token, token_hash, created_at, updated_at, expiry FROM tokens WHERE token = $1`

	var token Token

	row := db.QueryRowContext(ctx, query, plainText)
	err := row.Scan(
		&token.ID,
		&token.UserId,
		&token.Email,
		&token.Token,
		&token.TokenHash,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.Expiry,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (t *Token) GetUserForToken(token Token) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, first_name, last_name, password, user_active, created_at, updated_at FROM users WHERE id = $1`

	var user User
	row := db.QueryRowContext(ctx, query, &token.UserId)
	err := row.Scan(
		user.ID,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Password,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) { // TTL - time to life
	token := &Token{
		UserId: userID,
		Expiry: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Token = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Token))
	token.TokenHash = hash[:]

	return token, nil
}

func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, errors.New("no authorization header received")
	}

	headerParts := strings.Split(authorizationHeader, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no valid authorization header received")
	}

	token := headerParts[1]
	if len(token) != 26 {
		return nil, errors.New("token wrong size")
	}

	tfd, err := t.GetByToken(token)
	if err != nil {
		return nil, errors.New("no matching token found")
	}

	if tfd.Expiry.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	user, err := t.GetUserForToken(*tfd)
	if err != nil {
		return nil, errors.New("no matching user found")
	}

	if user.Active == false {
		return nil, errors.New("user is not active")
	}

	return user, nil
}

func (t *Token) Insert(token Token, user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `SELECT FROM tokens WHERE user_id = $1`

	_, err := db.ExecContext(ctx, stmt, token.UserId)
	if err != nil {
		return err
	}

	token.Email = user.Email

	stmt = `INSERT INTO tokens (user_id, email, token, token_hash, created_at, updated_at, expiry)
			VALUES($1, $2, $3, $4, $5, $6, $7)`

	_, err = db.ExecContext(ctx, stmt,
		token.UserId,
		token.Email,
		token.Token,
		token.TokenHash,
		token.CreatedAt,
		token.UpdatedAt,
		token.Expiry,
	)
	if err != nil {
		return err
	}

	return nil
}

func (t *Token) DeleteByToken(plainText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM tokens WHERE token = $1`

	_, err := db.ExecContext(ctx, stmt, plainText)
	if err != nil {
		return err
	}

	return nil
}

func (t *Token) DeleteTokensForUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM tokens WHERE user_id = $1`
	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

func (t *Token) ValidToken(plainText string) (bool, error) {
	token, err := t.GetByToken(plainText)
	if err != nil {
		return false, errors.New("no matching token found")
	}

	_, err = t.GetUserForToken(*token)
	if err != nil {
		return false, errors.New("no matching user found")
	}

	if token.Expiry.Before(time.Now()) {
		return false, errors.New("token has expired")
	}

	return true, nil
}

//SQL QUERY FOR A NEW PROJECT
//create table "users"
//(
//id          integer generated by default as identity,
//email       varchar(255) not null,
//first_name  varchar(255),
//last_name   varchar(255),
//password    varchar(69),
//created_at  timestamp(6) not null,
//updated_at  timestamp(6),
//user_active bool         not null
//);
//create table token
//(
//id         integer generated always as identity,
//user_id    integer      not null,
//email      varchar(255) not null,
//token      varchar(255) not null,
//token_hash bytea,
//created_at timestamp(6) not null,
//updated_at timestamp(6),
//expiry     timestamp(6)
//);
//alter table users
//add constraint key_name
//primary key (id);
//alter table token
//add constraint foreign_key_name
//foreign key (user_id) references users ("id");
