package rank

import (
	"cmp"
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewStore(rdb)
}

func TestRecordKill_OrdersByWin(t *testing.T) {
	s := testStore(t)
	ctx := t.Context()

	require.NoError(t, s.PutBasicInfo(ctx, 1, "alice", "Hero_A"))
	require.NoError(t, s.PutBasicInfo(ctx, 2, "bob", "Hero_B"))
	require.NoError(t, s.RecordKill(ctx, 2, 1))
	require.NoError(t, s.RecordKill(ctx, 2, 1))
	require.NoError(t, s.RecordKill(ctx, 1, 2))

	got, err := s.Top(ctx, 10)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, Item{RankID: 1, UserID: 2, UserName: "bob", HeroAvatar: "Hero_B", Win: 2}, got[0])
	assert.Equal(t, Item{RankID: 2, UserID: 1, UserName: "alice", HeroAvatar: "Hero_A", Win: 1}, got[1])
}

func TestTop_SkipsMissingBasicInfo(t *testing.T) {
	s := testStore(t)
	ctx := t.Context()

	require.NoError(t, s.RecordKill(ctx, 9, 8))

	got, err := s.Top(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestTop_RespectsLimit(t *testing.T) {
	s := testStore(t)
	ctx := t.Context()

	for id := range 12 {
		uid := id + 1
		require.NoError(t, s.PutBasicInfo(ctx, uid, "u", "Hero"))
		require.NoError(t, s.RecordKill(ctx, uid, 0))
	}

	got, err := s.Top(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, got, 10)
	assert.Equal(t, 1, got[0].RankID)
	assert.Equal(t, 10, got[len(got)-1].RankID)
}

// Hits the Redis the process would use in InitApp. Skips when nothing is listening
// so CI without Redis still runs the miniredis cases. Uses high ids and cleans up
// so a local Rank zset used by the game is not left with test members.
func TestLiveRedis_Smoke(t *testing.T) {
	addr := cmp.Or(os.Getenv("REDIS_ADDR"), "127.0.0.1:6379")
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Skipf("redis not available at %s: %v", addr, err)
	}

	const winner, loser = 9_000_001, 9_000_002
	t.Cleanup(func() {
		c := context.Background()
		_ = rdb.Del(c, userKey(winner), userKey(loser)).Err()
		_ = rdb.ZRem(c, rankKey, strconv.Itoa(winner), strconv.Itoa(loser)).Err()
	})

	s := NewStore(rdb)
	ctx := t.Context()
	require.NoError(t, s.PutBasicInfo(ctx, winner, "live-alice", "Hero_Shaman"))
	require.NoError(t, s.PutBasicInfo(ctx, loser, "live-bob", "Hero_Knight"))
	require.NoError(t, s.RecordKill(ctx, winner, loser))

	raw, err := rdb.HGet(ctx, userKey(winner), basicInfoField).Result()
	require.NoError(t, err)
	assert.Contains(t, raw, `"userId":9000001`)
	assert.Contains(t, raw, `"userName":"live-alice"`)
	assert.Contains(t, raw, `"heroAvatar":"Hero_Shaman"`)

	win, err := rdb.HGet(ctx, userKey(winner), winField).Int()
	require.NoError(t, err)
	assert.Equal(t, 1, win)

	lose, err := rdb.HGet(ctx, userKey(loser), loseField).Int()
	require.NoError(t, err)
	assert.Equal(t, 1, lose)

	score, err := rdb.ZScore(ctx, rankKey, strconv.Itoa(winner)).Result()
	require.NoError(t, err)
	assert.Equal(t, 1.0, score)

	_, err = s.Top(ctx, 10)
	require.NoError(t, err)
}

func TestOpen_RoundTrip(t *testing.T) {
	mr := miniredis.RunT(t)
	require.NoError(t, Open(mr.Addr()))
	t.Cleanup(func() {
		s := opened.Swap(nil)
		if s == nil {
			return
		}
		if c, ok := s.rdb.(*redis.Client); ok {
			_ = c.Close()
		}
	})

	ctx := t.Context()
	require.NoError(t, PutBasicInfo(ctx, 1, "alice", "Hero_A"))
	require.NoError(t, RecordKill(ctx, 1, 2))

	got, err := Top(ctx, 0)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, Item{RankID: 1, UserID: 1, UserName: "alice", HeroAvatar: "Hero_A", Win: 1}, got[0])
}
