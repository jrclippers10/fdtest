package main

import (
    "net/http"
    "io/ioutil"
	"math/rand"
    "errors"
    "time"
    "log"
    "sort"
    "strings"
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
    posCounts map[string]int
    teamCounts map[string]int
    Players []*Player
    Salary float64
    projectedScore float64
    status string
}

type ByProjectedScore []*FDLineup
func (s ByProjectedScore) Len() int {
    return len(s)
}
func (s ByProjectedScore) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByProjectedScore) Less(i, j int) bool {
    return s[i].projectedScore > s[j].projectedScore
}

type ByValue []Player
func (s ByValue) Len() int {
    return len(s)
}
func (s ByValue) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByValue) Less(i, j int) bool {
    return s[i].value > s[j].value
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
    log.Println(l.projectedScore)
    log.Println(l.Salary)
    for _,p := range l.Players {
        // fmt.Printf("%+v", p)
        log.Println(p.FirstName, p.LastName)

    }
    log.Println("**********")
    return
}

type FDPlayersResponse struct {
    Players []Player
}

var positions map[string][]Player

func newFDPlayersResponse(body []byte) *FDPlayersResponse {
    j := FDPlayersResponse{}
    _ = json.Unmarshal(body, &j)
    return &j
}

func (l *FDLineup) isValid() error {
    if len(l.Players) > 9 {
        return errors.New("Too many Players")
    }
    if l.posCounts["lw"] > 2 {
        return errors.New("Too many RWs")
    }
    if l.posCounts["rw"] > 2 {
        return errors.New("Too many LWs")
    }
    if l.posCounts["c"] > 2 {
        return errors.New("Too many Cs")
    }
    if l.posCounts["d"] > 2 {
        return errors.New("Too many Ds")
    }
    if l.posCounts["g"] > 1 {
        return errors.New("Too many Gs")
    }
    if l.Salary > 55000 {
        return errors.New("Invalid Salary")
    }
    for _, v := range l.teamCounts {
        if v > 4 {
            return errors.New("Too many players on one team")
        }
    }
    if len(l.Players) == 9 {
        l.status = "complete"
        for _, i := range l.Players {
            l.projectedScore += i.Fppg
        }
    }
    return nil
}

func (l *FDLineup) AddPlayer(p *Player) {
    if len(l.Players) >= 9 {
        return 
    }
    l.Players = append(l.Players, p)
    l.Salary += p.Salary
    l.posCounts[p.Position] += 1
    if _, ok := l.teamCounts[p.Team.Members[0]]; ok {
        l.teamCounts[p.Team.Members[0]] += 1
    } else {
        l.teamCounts[p.Team.Members[0]] = 1
    }
    if err := l.isValid(); err != nil {
        //delete last element
        lastI := len(l.Players) - 1
        l.Players, l.Players[lastI] = append(l.Players[:lastI], l.Players[lastI+1:]...), nil
        l.Salary -= p.Salary
        l.posCounts[strings.ToLower(p.Position)] -= 1
        l.teamCounts[p.Team.Members[0]] -= 1
    }
    return
}

func newFDLineup() *FDLineup{
    return &FDLineup{
        status: "new",
        posCounts: map[string]int{
            "RW": 0,
            "LW": 0,
            "C": 0,
            "D": 0,
            "G": 0,
        },
        teamCounts: map[string]int{},
        projectedScore: 0,
    }
}

func generateAllFDLineups(available []Player) []*FDLineup {
    var validLineups []*FDLineup
    for len(validLineups) < 150 {
        l := newFDLineup()
        for l.status == "new" {
            rn := rand.Intn(len(available)-1)
            p := available[rn]
            l.AddPlayer(&p)
        }
        if l.status == "complete" {
            l.PrintMe()
            time.Sleep(2 * time.Second)
            // validLineups = append(validLineups, l)
        }
    }
    return validLineups
}

func init() {
    positions = map[string][]Player{
        "RW": []Player{},
        "LW": []Player{},
        "D": []Player{}, 
        "C": []Player{}, 
        "G": []Player{},
    }
}

func main() {
	goGetIt()
}

func sortIntoGroups(available []Player) {
    for _, p := range available {
        positions[p.Position] = append(positions[p.Position], p)
    }
    return
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

func excludeLowValue(available []Player) []Player {
    ret := []Player{}
    for _, p := range available {
        if p.value > 4 {
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

func goGetIt() error {
	url := "https://api.fanduel.com/fixture-lists/14112/players"
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
    d := excludeLowScorers(b.Players)
    d = calcValues(d)
    log.Println("", len(d))
    d = excludeBadGoalies(d)
    log.Println("", len(d))
    d = excludeLowValue(d)
    log.Println("", len(d))
    // sort.Sort(ByValue(d))
    // sortIntoGroups(d)
    
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
