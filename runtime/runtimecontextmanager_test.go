//go:build !noquotas
// +build !noquotas

package runtime

import (
	"testing"
)

func Test_runtimeContextManager_LinearUnused(t *testing.T) {
	type fields struct {
		hardLimits    RuntimeResources
		usedResources RuntimeResources
	}
	type args struct {
		cpuFactor uint64
	}
	tests := []struct {
		name     string
		fields   fields
		trackCpu bool
		args     args
		want     uint64
	}{
		{
			name: "no limits",
			args: args{cpuFactor: 10},
			want: 0,
		},
		{
			name: "no mem limit",
			fields: fields{
				hardLimits: RuntimeResources{Cpu: 1000},
			},
			args: args{cpuFactor: 5},
			want: 5000,
		},
		{
			name: "no cpu limit",
			fields: fields{
				hardLimits: RuntimeResources{Memory: 1000},
			},
			args: args{cpuFactor: 5},
			want: 1000,
		},
		{
			name: "cpu wins",
			fields: fields{
				hardLimits: RuntimeResources{Memory: 1000, Cpu: 100},
			},
			args: args{cpuFactor: 5},
			want: 500,
		},
		{
			// A time limit counts CPU without a CPU limit: the budget is large
			// and finite (it used to wrap around: 0 - used).
			name: "cpu counted without cpu limit",
			fields: fields{
				hardLimits:    RuntimeResources{Millis: 1000},
				usedResources: RuntimeResources{Cpu: 12345},
			},
			trackCpu: true,
			args:     args{cpuFactor: 10},
			want:     untrackedCPUBudget * 10,
		},
		{
			name: "cpu counted without cpu limit, mem limit",
			fields: fields{
				hardLimits:    RuntimeResources{Millis: 1000, Memory: 1000},
				usedResources: RuntimeResources{Cpu: 12345, Memory: 100},
			},
			args: args{cpuFactor: 10},
			want: 900,
		},
		{
			name: "huge cpu limit fits in int64",
			fields: fields{
				hardLimits: RuntimeResources{Cpu: 1 << 62},
			},
			args: args{cpuFactor: 10},
			want: 1<<63 - 1,
		},
		{
			name: "mem wins",
			fields: fields{
				hardLimits: RuntimeResources{Memory: 1000, Cpu: 500},
			},
			args: args{cpuFactor: 10},
			want: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &runtimeContextManager{
				hardLimits:    tt.fields.hardLimits,
				usedResources: tt.fields.usedResources,
				trackCpu:      tt.trackCpu,
			}
			if got := m.LinearUnused(tt.args.cpuFactor); got != tt.want {
				t.Errorf("runtimeContextManager.LinearUnused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runtimeContextManager_TerminateContext(t *testing.T) {
	type fields struct {
		status RuntimeContextStatus
	}
	tests := []struct {
		name      string
		fields    fields
		wantPanic bool
	}{
		{
			name:      "live context",
			fields:    fields{status: StatusLive},
			wantPanic: true,
		},
		{
			name:      "done context",
			fields:    fields{status: StatusDone},
			wantPanic: false,
		},
		{
			name:      "error context",
			fields:    fields{status: StatusError},
			wantPanic: false,
		},
		{
			name:      "killed context",
			fields:    fields{status: StatusKilled},
			wantPanic: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &runtimeContextManager{
				status: tt.fields.status,
			}
			defer func() {
				switch recover() {
				case nil:
					if tt.wantPanic {
						t.Error("should panic")
					}
				default:
					if !tt.wantPanic {
						t.Error("should not panic")
					} else if m.status != StatusKilled {
						t.Error("context status should be StatusKilled")
					}
				}
			}()
			m.TerminateContext("error")
		})
	}
}

func Test_runtimeContextManager_UnusedWithoutLimit(t *testing.T) {
	m := &runtimeContextManager{
		hardLimits:    RuntimeResources{Millis: 1000},
		usedResources: RuntimeResources{Cpu: 12345, Memory: 678},
	}
	if got := m.UnusedCPU(); got != 0 {
		t.Errorf("UnusedCPU() = %d, want 0 (no CPU limit, CPU not counted)", got)
	}
	m.trackCpu = true
	if got := m.UnusedCPU(); got != untrackedCPUBudget {
		t.Errorf("UnusedCPU() = %d, want %d (CPU counted for the time limit)", got, uint64(untrackedCPUBudget))
	}
	m.trackCpu = false
	if got := m.UnusedMem(); got != 0 {
		t.Errorf("UnusedMem() = %d, want 0 (no memory limit)", got)
	}
	m.hardLimits.Cpu = 100
	m.usedResources.Cpu = 100
	if got := m.UnusedCPU(); got != 1 {
		t.Errorf("UnusedCPU() at the limit = %d, want 1", got)
	}
}
