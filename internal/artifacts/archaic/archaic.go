package archaic

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Set struct {
	element attributes.Element
	Index   int
	Count   int
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(core *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{Count: count}

	if count >= 2 {
		m := 0.15
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBase("archaic-2pc", -1),
			AffectedStat: attributes.GeoP,
			Amount: func() float64 {
				return m
			},
		})
	}
	if count >= 4 {
		enableSet := func(e attributes.Element) {
			s.element = e
			// Activate
			// TODO: cd for proc?
			core.Log.NewEvent("archaic petra proc'd", glog.LogArtifactEvent, char.Index()).
				Write("ele", s.element)

			// Apply mod to all characters
			for _, c := range core.Player.Chars() {
				c.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag("archaic-4pc", 10*60),
					AffectedStat: attributes.EleToDmgP(s.element),
					Amount: func() float64 {
						return 0.35
					},
				})
			}
		}

		core.Events.Subscribe(event.OnShielded, func(args ...any) {
			// Character that picks it up must be the petra set holder
			if core.Player.Active() != char.Index() {
				return
			}

			// Check shield
			shd := args[0].(shield.Shield)
			if shd.Type() != shield.Crystallize {
				return
			}
			enableSet(shd.Element())
		}, fmt.Sprintf("archaic-4pc-%v", char.Base.Key.String()))

		core.Events.Subscribe(event.OnLunarCrystallize, func(args ...any) {
			// Character that triggers it up must be the petra set holder
			if core.Player.Active() != char.Index() {
				return
			}
			enableSet(attributes.Hydro)
		}, fmt.Sprintf("archaic-4pc-%v", char.Base.Key.String()))
	}

	return &s, nil
}
