package main

import (
    "net/http"
    "io/ioutil"
    "errors"
    "log"
    "sort"
    "sync"
	  "encoding/json"
)

type Player struct {
    Salary float64 `json:"salary"`
    Injured bool `json:"injured"`
    FirstName string `json:"first_name"`
    LastName string `json:"last_name"`
    Played int `json:"played"`
    InjuryStatus interface{} `json:"injury_status"`
    Team struct {
        Members []string `json:"_members"`
        Ref string `json:"_ref"`
    } `json:"team"`
    Position string `json:"position"`
    Fppg float64 `json:"fppg"`
    Removed bool `json:"removed"`
    ID string `json:"id"`
    value float64
}

type FDLineup struct {
    Players []Player
    Salary float64
    ProjectedScore float64
}  

func (p *Player) String() (*string, error) {
    b, err := json.Marshal(p)
    if err != nil {
        return nil, err
    }
    s := string(b)
    return &s, nil
}

func (l *FDLineup) String() (*string, error) {
    b, err := json.Marshal(l)
    if err != nil {
        return nil, err
    }
    s := string(b)
    return &s, nil
}

func (l *FDLineup) PrintMe() {
    log.Println(l.ProjectedScore)
    log.Println(l.Salary)
    for _, p := range l.Players {
        // fmt.Printf("%+v", p)
        log.Println(p.FirstName, p.LastName)

    }
    log.Println("**********")
    return
}

type FDPlayersResponse struct {
    Players []Player
}

var positions = map[string][]Player{
        "RW": []Player{},
        "LW": []Player{},
        "D": []Player{}, 
        "C": []Player{}, 
        "G": []Player{},
    }
var lineupChan = make(chan FDLineup)
var validLineups []FDLineup


func newFDPlayersResponse(body []byte) *FDPlayersResponse {
    j := FDPlayersResponse{}
    _ = json.Unmarshal(body, &j)
    return &j
}

func (l *FDLineup) isValid() bool {
    if len(l.Players) > 9 {
        return false //errors.New("Too many Players")
    }
    if len(l.Players) < 9 {
        return false //errors.New("Too few Players")
    }
    if l.Salary > 55000 {
        return false //errors.New("Invalid Salary")
    }
    // for _, v := range l.teamCounts {
    //     if v > 4 {
    //         return errors.New("Too many players on one team")
    //     }
    // }
    return true
}

func (l *FDLineup) calcProjectedScore() {
    for _, p := range l.Players {
        l.ProjectedScore += p.Fppg
    }
    return
}

func (l *FDLineup) calcSalary() {
    for _, p := range l.Players {
        l.Salary += p.Salary
    }
    return
}

func newFDLineup(p []Player) FDLineup {
    players := make([]Player, 9)
    copy(players, p)
    l := FDLineup{
        Players: players,
    }
    l.calcSalary()
    l.calcProjectedScore()
    return l
}

// func init() {
// }

func pickNext(p string) (Player, error) {
    validPositions := []string{"RW", "LW", "C", "D", "G"}
    if !contains(validPositions, p) {
      return Player{}, errors.New("Invalid position requested")
    }
    // This does nothing at the moment
    return Player{}, nil
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func main() {
  var wg sync.WaitGroup
  wg.Add(1)

  go func() {
      validLineups = []FDLineup{}
      for len(validLineups) < 5000 {
          l := <- lineupChan
          if l.isValid() {
            validLineups = append(validLineups, l)
          }

      }
      sort.Sort(ByProjectedScore(validLineups))
      i := 0
      for i < 25 {
          validLineups[i].PrintMe()
          i++
      }
      wg.Done()
  }()
	goGetIt()
  wg.Wait()
}

func sortIntoGroups(available []Player) {
    for _, p := range available {
        positions[p.Position] = append(positions[p.Position], p)
    }
    return
}

func flattenGroups() []Player {
    ret := []Player{}
    for _, g := range positions {
        ret = append(ret, g...)
    }
    return ret
}

func excludeLowScorers(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        if p.Fppg > 2 {
            ret = append(ret, p)
        }
    }
    return ret
}

func excludeBadGoalies(available []Player) []Player {
    for i, p := range available {
        if p.Position == "G" && p.Fppg <= 3 {
            available = append(available[:i], available[i+1:]...)
        }
    }
    return available
}

func excludeInjured(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        if !p.Injured {
            ret = append(ret, p)
        }
    }
    return ret
}

func excludeLowValue(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        if p.value > 4 {
            ret = append(ret, p)
        }
    }
    return ret
}

func excludeLowGamesPlayed(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        if p.Played >= 5 {
            ret = append(ret, p)
        }
    }
    return ret
}

func calcValues(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        p.value = 10000*p.Fppg/p.Salary
        ret = append(ret, p)
    }
    return ret
}


func combine(v []Player, rwStart, lwStart, cStart, dStart, gStart, k, maxk int) {
    /* k here counts through positions in the maxk-element v.
     * if k > maxk, then the v is complete and we can use it.
     */
    if (k >= maxk) {
        l := newFDLineup(v)
        lineupChan <- l
        return
    }
    /* for this k'th element of the v, try all start..n
     * elements in that position
     */
    if k < 1 {
      for gStart < len(positions["G"]) {
        v[k] = positions["G"][gStart]
        /* recursively generate combinations of players
         * from i+1..n
         */
        combine(v, rwStart, lwStart, cStart, dStart, gStart+1, k+1, maxk)
        gStart++
      }
    } else if k < 3 {
      for rwStart < len(positions["RW"]) {
        v[k] = positions["RW"][rwStart]
        /* recursively generate combinations of players
         * from i+1..n
         */
        combine(v, rwStart+1, lwStart, cStart, dStart, gStart, k+1, maxk)
        rwStart++
      }
    } else if k < 5 {
      for lwStart < len(positions["LW"]) {
        v[k] = positions["LW"][lwStart]
        /* recursively generate combinations of players
         * from i+1..n
         */
        combine(v, rwStart, lwStart+1, cStart, dStart, gStart, k+1, maxk)
        lwStart++
      }
    } else if k < 7 {
      for cStart < len(positions["C"]) {
        v[k] = positions["C"][cStart]
        /* recursively generate combinations of players
         * from i+1..n
         */
        combine(v, rwStart, lwStart, cStart+1, dStart, gStart, k+1, maxk)
        cStart++
      }
    } else if k < 9 {
      for dStart < len(positions["D"]) {
        v[k] = positions["D"][dStart]
        /* recursively generate combinations of players
         * from i+1..n
         */
        combine(v, rwStart, lwStart, cStart, dStart+1, gStart, k+1, maxk)
        dStart++
      }
    }
}

func goGetIt() error {
	url := "https://api.fanduel.com/fixture-lists/14171/players"
    log.Println("URL: ", url)

    req, err := http.NewRequest("GET", url, nil)
    // req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Authorization", "Basic N2U3ODNmMTE4OTIzYzE2NzVjNWZhYWFmZTYwYTc5ZmM6")
    // req.Header.Set("X-Auth-Header", "2d04896f0d7b405f08a6fa3585e8c78b0b3d74335b60875c80b19b527a283a05")
    // req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    log.Println("response Status:", resp.Status)
    log.Println("response Headers:", resp.Header)
    body, _ := ioutil.ReadAll(resp.Body)
    b := newFDPlayersResponse(body)
    log.Println("", len(b.Players))
    d := excludeInjured(b.Players)
    d = excludeLowGamesPlayed(d)
    d = excludeBadGoalies(d)
    d = calcValues(d)
    sortIntoGroups(d)
    for i, g := range positions {
      var r []Player
      if i != "G" {
        sort.Sort(ByValue(g))
        r = make([]Player, 10)
        copy(r, g)
      } else {
        r = g
      }
      sort.Sort(ByFppg(r))
      g = make([]Player, 5)
      copy(g, r)
      positions[i] = g
    }


    // d = flattenGroups()
    // log.Println(d)
    // d = excludeLowValue(d)
    // log.Println("", len(d))


    /* generate all combinations of n elements taken
     * k at a time, starting with combinations containing 0
     * in the first position.
     */
    maxk := 9
    combine(make([]Player, maxk), 0, 0, 0, 0, 0, 0, maxk);
    close(lineupChan)

    // log.Println("goalies", len(positions["G"]))
    // for _, p := range positions["G"] {
    //     log.Println(p.FirstName, p.LastName, p.value)
    // }
    // a := generateAllFDLineups(b.Players)
    // sort.Sort(ByProjectedScore(a))
    // for i,element := range a {
    //     log.Println(i)
    //     element.PrintMe()
    // }
    // a[0].PrintMe()
    return nil
}
