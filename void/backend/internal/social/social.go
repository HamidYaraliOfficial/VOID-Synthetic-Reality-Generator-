// Package social implements VOID's Social Network Simulation: a dynamically
// growing follow/interaction graph supporting viral content propagation,
// echo-chamber formation, community detection (label propagation) and
// influencer-driven information cascades.
package social

import (
	"sync"

	"void/internal/rng"
)

// Graph is a directed follow graph plus per-post engagement tracking.
type Graph struct {
	mu        sync.RWMutex
	followers map[string]map[string]bool // target -> set of followers
	following map[string]map[string]bool // source -> set of followees
	posts     map[string]*Post
	communities map[string]int // entity id -> community label
}

type Post struct {
	ID       string
	AuthorID string
	Tick     int64
	Likes    int64
	Shares   int64
	Comments int64
	Reach    map[string]bool // set of entity ids that have seen it
}

func NewGraph() *Graph {
	return &Graph{
		followers:   map[string]map[string]bool{},
		following:   map[string]map[string]bool{},
		posts:       map[string]*Post{},
		communities: map[string]int{},
	}
}

func (g *Graph) Follow(source, target string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.following[source] == nil {
		g.following[source] = map[string]bool{}
	}
	if g.followers[target] == nil {
		g.followers[target] = map[string]bool{}
	}
	g.following[source][target] = true
	g.followers[target][source] = true
}

func (g *Graph) Unfollow(source, target string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.following[source], target)
	delete(g.followers[target], source)
}

func (g *Graph) FollowerCount(id string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.followers[id])
}

func (g *Graph) Publish(postID, authorID string, tick int64) *Post {
	g.mu.Lock()
	defer g.mu.Unlock()
	p := &Post{ID: postID, AuthorID: authorID, Tick: tick, Reach: map[string]bool{authorID: true}}
	g.posts[postID] = p
	return p
}

// Propagate simulates one round of information cascade: every entity who
// has seen the post may re-share it to their own followers, weighted by an
// influence score and a virality coefficient k (from a Scenario override).
func (g *Graph) Propagate(postID string, k float64, r *rng.Source) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	p, ok := g.posts[postID]
	if !ok {
		return 0
	}
	newlyReached := 0
	current := make([]string, 0, len(p.Reach))
	for id := range p.Reach {
		current = append(current, id)
	}
	for _, id := range current {
		for follower := range g.followers[id] {
			if p.Reach[follower] {
				continue
			}
			prob := 0.05 * k
			if r.Bool(prob) {
				p.Reach[follower] = true
				newlyReached++
				if r.Bool(0.3) {
					p.Likes++
				}
				if r.Bool(0.05) {
					p.Shares++
				}
			}
		}
	}
	return newlyReached
}

// DetectCommunities runs a simplified label-propagation pass to approximate
// community formation for the Network Graph visualization.
func (g *Graph) DetectCommunities(iterations int, r *rng.Source) map[string]int {
	g.mu.Lock()
	defer g.mu.Unlock()

	ids := make([]string, 0, len(g.following)+len(g.followers))
	seen := map[string]bool{}
	for id := range g.following {
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	for id := range g.followers {
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	for i, id := range ids {
		g.communities[id] = i
	}
	for iter := 0; iter < iterations; iter++ {
		for _, id := range ids {
			neighborLabels := map[int]int{}
			for f := range g.following[id] {
				neighborLabels[g.communities[f]]++
			}
			for f := range g.followers[id] {
				neighborLabels[g.communities[f]]++
			}
			best, bestCount := g.communities[id], -1
			for label, count := range neighborLabels {
				if count > bestCount {
					best, bestCount = label, count
				}
			}
			g.communities[id] = best
		}
	}
	out := make(map[string]int, len(g.communities))
	for k, v := range g.communities {
		out[k] = v
	}
	return out
}
