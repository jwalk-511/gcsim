package speed

// Assumes frames only ever increase or stay the same
// Assumes hitlag only ever occurs after previous hitlag has ended
type Speed struct {
	speed        float64 // Recorded in frames / s internally
	displacement float64 // Recorded in meters
	speedSrc     int     // Record the last time speed was updated, in game ticks (frames)
	hitlagDur    int     // How many frames of hitlag most recently occured
	hitlagSrc    int     // Which frame the most recent hitlag started
}

// Should only be called when iteration begins, so initial src frame is 0.
// If it should be called arbitrarily within an iteration, setSpd should be called immediately after to apply a src frame
// Speed here is given in m/s and converted to m/frame
func NewSpd(speed float64) *Speed {
	return &Speed{
		speed: speed / 60.0,
	}
}

// frame int: Current frame
// Return: how many frames of hitlag to subtract from the time elapsed, for calculating displacement
func (s *Speed) hitlagAdjDur(frame int) int {
	// If currently in hitlag, return how long have been in hitlag
	if s.hitlagSrc+s.hitlagDur > frame {
		return frame - s.hitlagSrc
	}

	return s.hitlagDur
}

// Returns speed in m/s
// For external use only
func (s *Speed) Spd() float64 {
	return s.speed * 60
}

// frame int: Current frame
func (s *Speed) Displacement(frame int) float64 {
	return s.displacement + (s.speed * float64(frame-s.speedSrc-s.hitlagAdjDur(frame)))
}

// Apply previous hitlag to displacement and set new hitlag
// Assumes previous hitlag must have ended before new hitlag is added
func (s *Speed) ApplyHitlag(dur int, src int) {
	s.displacement -= s.speed * float64(s.hitlagDur)
	s.hitlagDur = dur
	s.hitlagSrc = src
}

// speed float64: Speed of entity in m/s
// frame int: Current frame
//
// When setting a new speed:
//  1. Update the accumulated displacement based on old speed and src
//  2. Subtract out frozen time due to hitlag
//     (1) and (2) are accomplished within s.Displacement()
//  3. Update current speed and src
//  4. Update hitlag record keeping.
//     If hitlag could span multiple speed values, other calculations wouldn't function
func (s *Speed) SetSpd(speed float64, frame int) {
	s.displacement = s.Displacement(frame)

	// Convert m/s to m/frame
	s.speed = speed / 60.0
	s.speedSrc = frame

	s.hitlagDur = s.hitlagDur - s.hitlagAdjDur(frame)
	s.hitlagSrc = frame
}
