package store

import "github.com/lacsar712/pasteur/internal/model"

type CarrierSnapshotView struct {
	UnitID   string
	Carrier     model.CarrierReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneCarrierSnapshot(s model.PlantSnapshot) CarrierSnapshotView {
	out := CarrierSnapshotView{
		UnitID:   s.UnitID,
		Carrier:     s.Carrier,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
