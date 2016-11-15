package main

type ByProjectedScore []FDLineup
func (s ByProjectedScore) Len() int {
    return len(s)
}
func (s ByProjectedScore) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByProjectedScore) Less(i, j int) bool {
    return s[i].ProjectedScore > s[j].ProjectedScore
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

type ByFppg []Player
func (s ByFppg) Len() int {
    return len(s)
}
func (s ByFppg) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}
func (s ByFppg) Less(i, j int) bool {
    return s[i].Fppg > s[j].Fppg
}