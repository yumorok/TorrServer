package torrent

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"server/search/parser"

	"github.com/anacrolix/torrent/metainfo"
)

func Search(query string, filterStrings []string) []*parser.Torrent {
	var wa sync.WaitGroup
	var mu sync.Mutex
	list := make([]*parser.Torrent, 0)

	wa.Add(3)
	go func() {
		defer wa.Done()
		lst, err := parser.NewRutor().Search(query)
		fmt.Println("End rutor")
		if err != nil {
			fmt.Println("Rutor search err:", err)
			return
		}
		mu.Lock()
		list = append(list, lst...)
		mu.Unlock()
	}()

	go func() {
		defer wa.Done()
		lst, err := parser.NewYHH().Search(query)
		fmt.Println("End YHH")
		if err != nil {
			fmt.Println("Yohoho search err:", err)
			return
		}
		mu.Lock()
		list = append(list, lst...)
		mu.Unlock()
	}()

	go func() {
		defer wa.Done()
		lst, err := parser.NewTParser().Search(query)
		fmt.Println("End TParser")
		if err != nil {
			fmt.Println("TParser search err:", err)
			return
		}
		mu.Lock()
		list = append(list, lst...)
		mu.Unlock()
	}()
	wa.Wait()
	filterStrings = append(filterStrings, query)
	fmt.Println("Filtering...", filterStrings)
	start := time.Now()
	defer func() {
		fmt.Println("End filtering", time.Since(start).Seconds())
	}()
	return filter(list, filterStrings)
}

func filter(list []*parser.Torrent, filterStrings []string) []*parser.Torrent {
	filtered := make([]*parser.Torrent, 0)
	for _, t := range removeDublicate(list) {
		name := strings.ToLower(t.Name)
		var isFilter = false
		for _, f := range filterStrings {
			if f := strings.TrimSpace(strings.ToLower(f)); f != "" {
				if strings.Contains(f, "|") {
					ff := strings.Split(f, "|")
					isFound := false
					for _, fs := range ff {
						if strings.Contains(name, strings.TrimSpace(fs)) {
							isFound = true
							break
						}
					}
					if !isFound {
						isFilter = true
						break
					}

				} else {
					if !strings.Contains(name, f) {
						isFilter = true
						break
					}
				}
			}
		}
		if !isFilter {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func removeDublicate(list []*parser.Torrent) []*parser.Torrent {
	magnets := map[string]*parser.Torrent{}
	for _, t := range list {
		mag := getMagnet(t.Magnet)
		if torr, ok := magnets[mag.InfoHash.HexString()]; !ok {
			magnets[mag.InfoHash.HexString()] = t
		} else {
			smag := getMagnet(torr.Magnet)
			mag = concatMagnet(mag, smag)
			if mag.DisplayName != torr.Name {
				mag.DisplayName = torr.Name
			}
			torr.Magnet = mag.String()
			if torr.PeersDl < t.PeersDl {
				torr.PeersDl = t.PeersDl
			}
			if torr.PeersUl < t.PeersUl {
				torr.PeersUl = t.PeersUl
			}
			magnets[mag.InfoHash.HexString()] = torr
		}
	}

	torrents := []*parser.Torrent{}
	for _, t := range magnets {
		torrents = append(torrents, t)
	}
	return torrents
}

func getMagnet(magStr string) *metainfo.Magnet {
	m, err := metainfo.ParseMagnetURI(magStr)
	if err != nil {
		return nil
	}
	return &m
}

func concatMagnet(m1, m2 *metainfo.Magnet) *metainfo.Magnet {
	n1 := m1.DisplayName
	n2 := m2.DisplayName
	if len(n1) < len(n2) {
		n1 = n2
	}
	trackers := map[string]struct{}{}
	for _, tr := range m1.Trackers {
		trackers[tr] = struct{}{}
	}
	for _, tr := range m2.Trackers {
		trackers[tr] = struct{}{}
	}

	m1.Trackers = []string{}
	for tr := range trackers {
		m1.Trackers = append(m1.Trackers, tr)
	}
	m1.DisplayName = n1
	return m1
}
