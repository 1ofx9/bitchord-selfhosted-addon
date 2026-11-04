package library

import (
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/rairulyle/bitchord-selfhosted-addon/internal/media"
)

type Index struct {
	tracks      []media.Track
	byID        map[string]int
	postings    map[string][]int
	titlePosts  map[string][]int
	titleTokens [][]string
	titleForms  [][]string
}

func NewIndex(tracks []media.Track) *Index {
	ix := &Index{
		tracks:      tracks,
		byID:        make(map[string]int, len(tracks)),
		postings:    map[string][]int{},
		titlePosts:  map[string][]int{},
		titleTokens: make([][]string, len(tracks)),
		titleForms:  make([][]string, len(tracks)),
	}
	for i, t := range tracks {
		ix.byID[t.ID] = i
		title := unique(Tokens(t.Title))
		ix.titleTokens[i] = title
		for _, token := range title {
			ix.titlePosts[token] = append(ix.titlePosts[token], i)
		}
		everything := strings.Join([]string{t.Artist, t.AlbumArtist, t.Album}, " ")
		hits := map[string]struct{}{}
		for _, token := range title {
			hits[token] = struct{}{}
		}
		for _, form := range bitchordWords(t.Title) {
			if _, known := hits[form]; !known {
				ix.titleForms[i] = append(ix.titleForms[i], form)
			}
			hits[form] = struct{}{}
		}
		for _, token := range unique(Tokens(everything)) {
			hits[token] = struct{}{}
		}
		for _, form := range bitchordWords(everything) {
			hits[form] = struct{}{}
		}
		for token := range hits {
			ix.postings[token] = append(ix.postings[token], i)
		}
	}
	return ix
}

func (ix *Index) Len() int { return len(ix.tracks) }

func (ix *Index) missingFrom(other *Index) int {
	if other == nil {
		return len(ix.tracks)
	}
	missing := 0
	for id := range ix.byID {
		if _, ok := other.byID[id]; !ok {
			missing++
		}
	}
	return missing
}

func (ix *Index) Get(id string) (media.Track, bool) {
	i, ok := ix.byID[id]
	if !ok {
		return media.Track{}, false
	}
	return ix.tracks[i], true
}

type Result struct {
	Tracks   []media.Track
	Strict   int
	Fallback int
}

var noiseWords = map[string]struct{}{
	"live": {}, "remaster": {}, "remastered": {}, "feat": {}, "ft": {}, "featuring": {},
	"deluxe": {}, "edition": {}, "bonus": {}, "track": {}, "acoustic": {}, "instrumental": {},
	"official": {}, "video": {}, "audio": {}, "version": {}, "edit": {}, "mix": {}, "remix": {},
	"pt": {}, "part": {}, "vol": {}, "volume": {},
}

func isYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	for i := 0; i < 4; i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s >= "1900" && s <= "2099"
}

func stripNoise(q []string) []string {
	var cleaned []string
	for _, t := range q {
		if _, noise := noiseWords[t]; !noise && !isYear(t) {
			cleaned = append(cleaned, t)
		}
	}
	return cleaned
}

func (ix *Index) lookupPostings(token string, prefix bool) []int {
	exact := ix.postings[token]
	if len(exact) > 0 || !prefix || len(token) < 2 {
		return exact
	}
	var matched []int
	seen := make(map[int]struct{})
	for key, posts := range ix.postings {
		if strings.HasPrefix(key, token) {
			for _, id := range posts {
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					matched = append(matched, id)
				}
			}
		}
	}
	sort.Ints(matched)
	return matched
}

func (ix *Index) lookupTitle(token string, prefix bool) []int {
	exact := ix.titlePosts[token]
	if len(exact) > 0 || !prefix || len(token) < 2 {
		return exact
	}
	var matched []int
	seen := make(map[int]struct{})
	for key, posts := range ix.titlePosts {
		if strings.HasPrefix(key, token) {
			for _, id := range posts {
				if _, ok := seen[id]; !ok {
					seen[id] = struct{}{}
					matched = append(matched, id)
				}
			}
		}
	}
	sort.Ints(matched)
	return matched
}

func (ix *Index) Search(query string, limit int) []media.Track { return ix.Find(query, limit).Tracks }

func (ix *Index) Find(query string, limit int) Result {
	q := unique(Tokens(query))
	if len(q) == 0 || limit <= 0 {
		return Result{}
	}
	strict := ix.strict(q, false)
	seen := make(map[int]struct{}, len(strict))
	for _, i := range strict {
		seen[i] = struct{}{}
	}
	var loose []int
	for _, i := range ix.fallback(q, false) {
		if _, dup := seen[i]; !dup {
			loose = append(loose, i)
		}
	}

	// Tier 3: If no matches, attempt prefix matching for incomplete words
	if len(strict) == 0 && len(loose) == 0 {
		for _, i := range ix.strict(q, true) {
			strict = append(strict, i)
			seen[i] = struct{}{}
		}
		if len(strict) == 0 {
			for _, i := range ix.fallback(q, true) {
				if _, dup := seen[i]; !dup {
					loose = append(loose, i)
				}
			}
		}
	}

	// Tier 4: If still no matches, strip variant/noise tokens (live, remaster, feat, etc.)
	if len(strict) == 0 && len(loose) == 0 {
		cleaned := stripNoise(q)
		if len(cleaned) > 0 && len(cleaned) < len(q) {
			cleanStrict := ix.strict(cleaned, false)
			if len(cleanStrict) == 0 {
				cleanStrict = ix.strict(cleaned, true)
			}
			for _, i := range cleanStrict {
				strict = append(strict, i)
				seen[i] = struct{}{}
			}
			if len(strict) == 0 {
				cleanLoose := ix.fallback(cleaned, false)
				if len(cleanLoose) == 0 {
					cleanLoose = ix.fallback(cleaned, true)
				}
				for _, i := range cleanLoose {
					if _, dup := seen[i]; !dup {
						loose = append(loose, i)
					}
				}
			}
		}
	}

	ix.rank(strict, q)
	ix.rank(loose, q)
	result := Result{Strict: len(strict), Fallback: len(loose)}
	ordered := append(strict, loose...)
	if len(ordered) > limit {
		ordered = ordered[:limit]
	}
	out := make([]media.Track, len(ordered))
	for n, i := range ordered {
		out[n] = ix.tracks[i]
	}
	result.Tracks = out
	return result
}

func (ix *Index) strict(q []string, prefix bool) []int {
	lists := make([][]int, len(q))
	for n, token := range q {
		lists[n] = ix.lookupPostings(token, prefix)
		if len(lists[n]) == 0 {
			return nil
		}
	}
	sort.Slice(lists, func(a, b int) bool { return len(lists[a]) < len(lists[b]) })
	result := slices.Clone(lists[0])
	for _, next := range lists[1:] {
		result = intersect(result, next)
		if len(result) == 0 {
			return nil
		}
	}
	return result
}

func (ix *Index) fallback(q []string, prefix bool) []int {
	hits := map[int]int{}
	for _, token := range q {
		for _, i := range ix.lookupTitle(token, prefix) {
			hits[i]++
		}
	}
	var out []int
	for i, n := range hits {
		if count := len(ix.titleTokens[i]); count >= 2 && n >= count {
			out = append(out, i)
		}
	}
	return out
}

func (ix *Index) rank(candidates []int, q []string) {
	wanted := make(map[string]struct{}, len(q))
	for _, token := range q {
		wanted[token] = struct{}{}
	}
	titleHits := make(map[int]int, len(candidates))
	for _, i := range candidates {
		for _, tokens := range [][]string{ix.titleTokens[i], ix.titleForms[i]} {
			for _, token := range tokens {
				if _, ok := wanted[token]; ok {
					titleHits[i]++
				}
			}
		}
	}
	sort.Slice(candidates, func(a, b int) bool {
		ia, ib := candidates[a], candidates[b]
		ta, tb := ix.tracks[ia], ix.tracks[ib]
		if titleHits[ia] != titleHits[ib] {
			return titleHits[ia] > titleHits[ib]
		}
		if la, lb := utf8.RuneCountInString(ta.Title), utf8.RuneCountInString(tb.Title); la != lb {
			return la < lb
		}
		if ta.Album != tb.Album {
			return ta.Album < tb.Album
		}
		return ta.ID < tb.ID
	})
}

func intersect(a, b []int) []int {
	out := a[:0]
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, a[i])
			i++
			j++
		case a[i] < b[j]:
			i++
		default:
			j++
		}
	}
	return out
}

func unique(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := tokens[:0]
	for _, token := range tokens {
		if _, dup := seen[token]; !dup {
			seen[token] = struct{}{}
			out = append(out, token)
		}
	}
	return out
}
