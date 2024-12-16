//go:build !race

package version

// isRace выставление булевы в false, если сервер собран без флага race
const isRace = false
