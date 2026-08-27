// Package rank is the kill ladder. Redis key names follow the course
// (User_{id} hash, Rank zset) so existing ops notes stay valid.
//
// Store is the unit of access, like sql.DB. Open installs the process-wide
// instance; package functions then match (*Store) methods, same as
// http.ListenAndServe vs (*Server).ListenAndServe.
package rank

import (
	"context"
	"encoding/json/v2"
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	rankKey        = "Rank"
	winField       = "Win"
	loseField      = "Lose"
	basicInfoField = "BasicInfo"
	defaultTopN    = 10
)

// Item is one GetRankResult row.
type Item struct {
	RankID     int
	UserID     int
	UserName   string
	HeroAvatar string
	Win        uint32
}

type basicInfo struct {
	UserID     int    `json:"userId"`
	UserName   string `json:"userName"`
	HeroAvatar string `json:"heroAvatar"`
}

// Store talks to one Redis. Tests pass a miniredis client; Open stores the
// process-wide instance used by the package functions.
type Store struct {
	rdb redis.Cmdable
}

func NewStore(rdb redis.Cmdable) *Store {
	return &Store{rdb: rdb}
}

var opened atomic.Pointer[Store]

// Open pings addr and makes it the process store. InitApp fatals on error.
func Open(addr string) error {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		return err
	}
	if old := opened.Swap(NewStore(rdb)); old != nil {
		if c, ok := old.rdb.(*redis.Client); ok {
			_ = c.Close()
		}
	}
	return nil
}

func PutBasicInfo(ctx context.Context, userID int, userName, heroAvatar string) error {
	s := opened.Load()
	if s == nil {
		return nil
	}
	return s.PutBasicInfo(ctx, userID, userName, heroAvatar)
}

func RecordKill(ctx context.Context, winnerID, loserID int) error {
	s := opened.Load()
	if s == nil {
		return nil
	}
	return s.RecordKill(ctx, winnerID, loserID)
}

func Top(ctx context.Context, n int) ([]Item, error) {
	s := opened.Load()
	if s == nil {
		return nil, nil
	}
	return s.Top(ctx, n)
}

func userKey(userID int) string {
	return "User_" + strconv.Itoa(userID)
}

func (s *Store) PutBasicInfo(ctx context.Context, userID int, userName, heroAvatar string) error {
	if userID <= 0 {
		return nil
	}
	b, err := json.Marshal(&basicInfo{
		UserID:     userID,
		UserName:   userName,
		HeroAvatar: heroAvatar,
	})
	if err != nil {
		return err
	}
	return s.rdb.HSet(ctx, userKey(userID), basicInfoField, b).Err()
}

// RecordKill bumps the winner's Win (and Rank score) and the loser's Lose.
// A non-positive id is skipped so a bot kill still records the other side.
func (s *Store) RecordKill(ctx context.Context, winnerID, loserID int) error {
	if winnerID <= 0 && loserID <= 0 {
		return nil
	}
	if winnerID > 0 {
		win, err := s.rdb.HIncrBy(ctx, userKey(winnerID), winField, 1).Result()
		if err != nil {
			return err
		}
		if err := s.rdb.ZAdd(ctx, rankKey, redis.Z{
			Score:  float64(win),
			Member: strconv.Itoa(winnerID),
		}).Err(); err != nil {
			return err
		}
	}
	if loserID > 0 {
		if err := s.rdb.HIncrBy(ctx, userKey(loserID), loseField, 1).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Top(ctx context.Context, n int) ([]Item, error) {
	if n <= 0 {
		n = defaultTopN
	}

	tuples, err := s.rdb.ZRevRangeWithScores(ctx, rankKey, 0, int64(n-1)).Result()
	if err != nil {
		return nil, err
	}

	out := make([]Item, 0, len(tuples))
	for _, t := range tuples {
		member, ok := t.Member.(string)
		if !ok {
			continue
		}
		userID, err := strconv.Atoi(member)
		if err != nil {
			continue
		}
		raw, err := s.rdb.HGet(ctx, userKey(userID), basicInfoField).Bytes()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var info basicInfo
		if err := json.Unmarshal(raw, &info); err != nil {
			log.Warn().Err(err).Int("userId", userID).Msg("rank: skip bad BasicInfo")
			continue
		}
		out = append(out, Item{
			RankID:     len(out) + 1,
			UserID:     userID,
			UserName:   info.UserName,
			HeroAvatar: info.HeroAvatar,
			Win:        uint32(t.Score),
		})
	}
	return out, nil
}
