package torrent

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"server/search/parser"
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
	encountered := map[string]struct{}{}
	result := []*parser.Torrent{}

	for _, t := range list {
		if _, ok := encountered[getHash(t)]; !ok {
			encountered[getHash(t)] = struct{}{}
			result = append(result, t)
		}
	}
	return result
}

func getHash(t *parser.Torrent) string {
	return strings.ToLower(t.Magnet)
}
