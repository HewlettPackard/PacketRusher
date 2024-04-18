/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package scenario

import (
	"io"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var defaultUeId int = 1

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func getDefaultManager() ScenarioManager {
	s := ScenarioManager{
		ues: map[int]*ueTasksCtx{},
	}
	s.registrationQueue = make(chan order, 5)
	ue := &ueTasksCtx{}
	ue.workerChan = make(chan Task, 5)
	ue.nextTasks = []Task{}
	s.ues[defaultUeId] = ue
	s.taskExecutorWg = &sync.WaitGroup{}
	s.gnbWg = &sync.WaitGroup{}
	return s
}

// create gnb context  without starting services
func gnbInitMock(gnbConf config.GNodeB, amfConfs []*config.AMF, wg *sync.WaitGroup) *context.GNBContext {
	// instance new gnb.
	gnb := &context.GNBContext{}

	// new gnb context.
	gnb.NewRanGnbContext(
		gnbConf.PlmnList.GnbId,
		gnbConf.PlmnList.Mcc,
		gnbConf.PlmnList.Mnc,
		gnbConf.PlmnList.Tac,
		gnbConf.SliceSupportList.Sst,
		gnbConf.SliceSupportList.Sd,
		gnbConf.ControlIF.Ip,
		gnbConf.DataIF.Ip,
		gnbConf.ControlIF.Port,
		gnbConf.DataIF.Port)

	gnb.NewGnBAmf(amfConfs[0].Ip, amfConfs[0].Port)

	return gnb
}

func TestHandleReport(t *testing.T) {
	type input struct {
		tasks  []Task
		report report
	}

	type testCases struct {
		description string
		input       input
		expected    []Task
	}

	tasks := []Task{
		{
			TaskType: Registration,
			Delay:    2000,
		},
		{
			TaskType: Idle,
			Delay:    500,
		},
	}

	scenarios := []testCases{
		{
			description: "successful report should not change tasks",
			input: input{
				tasks: tasks,
				report: report{
					ueId:    defaultUeId,
					success: true,
					reason:  "sucessful",
				},
			},
			expected: tasks,
		},
		{
			description: "failed report should empty tasks list",
			input: input{
				tasks: tasks,
				report: report{
					ueId:    defaultUeId,
					success: false,
					reason:  "fail",
				},
			},
			expected: []Task{},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := &ScenarioManager{
				ues: map[int]*ueTasksCtx{defaultUeId: {nextTasks: tasks}},
			}
			s.handleReport(scenario.input.report)
			assert.Equal(t, scenario.expected, s.ues[defaultUeId].nextTasks)
		})
	}
}

func TestManageNextTask(t *testing.T) {
	type input struct {
		manager    ScenarioManager
		tasks      []Task
		UeScenario UEScenario
	}

	type expected struct {
		nextTasks []Task
		err       bool
	}

	type testCases struct {
		description string
		input       input
		expected    expected
	}

	initialTasks := []Task{{
		TaskType: AttachToGNB,
		Delay:    2000,
	}, {
		TaskType: Registration,
		Delay:    2000,
	}, {
		TaskType: Idle,
		Delay:    500,
	}}

	scenarios := []testCases{{
		description: "Should send next task",
		input: input{
			tasks:   initialTasks,
			manager: getDefaultManager(),
			UeScenario: UEScenario{
				Tasks: initialTasks,
				Loop:  false,
			},
		},
		expected: expected{
			nextTasks: []Task{{
				TaskType: Registration,
				Delay:    2000,
			}, {
				TaskType: Idle,
				Delay:    500,
			}},
			err: false,
		},
	}, {
		description: "Should send next task without restarting scenario",
		input: input{
			tasks:   initialTasks,
			manager: getDefaultManager(),
			UeScenario: UEScenario{
				Tasks: initialTasks,
				Loop:  true,
			},
		},
		expected: expected{
			nextTasks: []Task{{
				TaskType: Registration,
				Delay:    2000,
			}, {
				TaskType: Idle,
				Delay:    500,
			}},
			err: false,
		},
	}, {
		description: "Should restart scenario",
		input: input{
			tasks:   []Task{},
			manager: getDefaultManager(),
			UeScenario: UEScenario{
				Tasks: initialTasks,
				Loop:  true,
			},
		},
		expected: expected{
			nextTasks: initialTasks,
			err:       false,
		},
	}, {
		description: "Should return error",
		input: input{
			tasks:   []Task{},
			manager: getDefaultManager(),
			UeScenario: UEScenario{
				Tasks: initialTasks,
				Loop:  false,
			},
		},
		expected: expected{
			nextTasks: []Task{},
			err:       true,
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := scenario.input.manager
			s.ues[defaultUeId].nextTasks = scenario.input.tasks
			err := s.manageNextTask(defaultUeId, scenario.input.UeScenario)
			assert.Equal(t, scenario.expected.nextTasks, s.ues[defaultUeId].nextTasks)
			if scenario.expected.err {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestSendNextTask(t *testing.T) {
	type input struct {
		manager ScenarioManager
		ueid    int
		tasks   []Task
	}

	type expected struct {
		nextTasks    []Task
		sentToWorker Task
		sentToQueue  order
		err          bool
	}

	type testCases struct {
		description string
		input       input
		expected    expected
	}

	manager := getDefaultManager()

	scenarios := []testCases{
		{
			description: "task is send to worker",
			input: input{
				tasks: []Task{{
					TaskType: AttachToGNB,
					Delay:    2000,
				}, {
					TaskType: Registration,
					Delay:    2000,
				}, {
					TaskType: Idle,
					Delay:    500,
				}},
				manager: getDefaultManager(),
				ueid:    defaultUeId,
			},
			expected: expected{
				nextTasks: []Task{{
					TaskType: Registration,
					Delay:    2000,
				}, {
					TaskType: Idle,
					Delay:    500,
				}},
				sentToWorker: Task{
					TaskType: AttachToGNB,
					Delay:    2000,
				},
				sentToQueue: order{},
				err:         false,
			},
		},
		{
			description: "registration task is send to queue for timer handling",
			input: input{
				tasks: []Task{{
					TaskType: Registration,
					Delay:    2000,
				}, {
					TaskType: Idle,
					Delay:    500,
				}},
				manager: manager,
				ueid:    defaultUeId,
			},
			expected: expected{
				nextTasks: []Task{{
					TaskType: Idle,
					Delay:    500,
				}},
				sentToWorker: Task{},
				sentToQueue: order{
					workerChan: manager.ues[defaultUeId].workerChan,
					task: Task{
						TaskType: Registration,
						Delay:    2000,
					},
				},
				err: false,
			},
		},
		{
			description: "Does not have task left",
			input: input{
				tasks:   []Task{},
				manager: getDefaultManager(),
				ueid:    defaultUeId,
			},
			expected: expected{
				nextTasks:    []Task{},
				sentToWorker: Task{},
				sentToQueue:  order{},
				err:          true,
			},
		},
		{
			description: "Ue not found",
			input: input{
				tasks:   []Task{},
				manager: getDefaultManager(),
				ueid:    2,
			},
			expected: expected{
				nextTasks:    []Task{},
				sentToWorker: Task{},
				sentToQueue:  order{},
				err:          true,
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {

			s := scenario.input.manager
			s.ues[defaultUeId].nextTasks = scenario.input.tasks
			res := s.sendNextTask(defaultUeId)

			var err string

			if scenario.expected.sentToWorker != (Task{}) {
				select {
				case task := <-s.ues[defaultUeId].workerChan:
					assert.Equal(t, scenario.expected.sentToWorker, task)
				default:
					t.Error("Task expected to be sent to worker but was not.")
				}
				err = "More than one task was send to worker"
			} else {
				err = "Worker received a task but was not supposed to"
			}
			select {
			case <-s.ues[defaultUeId].workerChan:
				t.Error(err)
			default:
			}

			if scenario.expected.sentToQueue != (order{}) {
				select {
				case task := <-s.registrationQueue:
					assert.Equal(t, scenario.expected.sentToQueue, task)
				default:
					t.Error("Task expected to be sent to registration queue but was not.")
				}
				err = "More than one task was send to registration queue"
			} else {
				err = "Registration queue received a task but was not supposed to"
			}
			select {
			case <-s.registrationQueue:
				t.Error(err)
			default:
			}

			if scenario.expected.err {
				assert.Error(t, res)
			} else {
				assert.Nil(t, res, "SendNextTask returned an error but was not supposed to")
			}

			assert.Equal(t, scenario.expected.nextTasks, s.ues[defaultUeId].nextTasks)
		})
	}

}

func TestRestartUeScenario(t *testing.T) {
	type input struct {
		manager ScenarioManager
		tasks   []Task
	}

	type testCases struct {
		description string
		input       input
		expected    []Task
	}

	tasks := []Task{{
		TaskType: AttachToGNB,
		Delay:    0,
	}, {
		TaskType: Registration,
		Delay:    2000,
	}, {
		TaskType: NewPDUSession,
		Delay:    0,
	}, {
		TaskType: Idle,
		Delay:    500,
	}}

	scenarios := []testCases{{
		description: "UE should be killed and tasks reset",
		input: input{
			manager: getDefaultManager(),
			tasks: []Task{{
				TaskType: Idle,
				Delay:    500,
			}},
		},
		expected: tasks,
	}, {
		description: "UE should be killed and tasks reset",
		input: input{
			manager: getDefaultManager(),
			tasks:   []Task{},
		},
		expected: tasks,
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := scenario.input.manager
			ueScenario := UEScenario{
				Tasks: tasks,
			}
			s.restartUeScenario(defaultUeId, ueScenario)

			select {
			case task := <-s.ues[defaultUeId].workerChan:
				assert.Equal(t, Task{TaskType: Kill}, task)
			default:
				t.Error("Kill Task expected to be sent to worker but was not.")
			}
			assert.Equal(t, scenario.expected, s.ues[defaultUeId].nextTasks)
		})
	}
}

func TestRestartUeScenarioHasErrors(t *testing.T) {

	type input struct {
		manager ScenarioManager
		tasks   []Task
		ueId    int
	}

	type testCases struct {
		description string
		input       input
	}

	scenarios := []testCases{{
		description: "Ue does not exist",
		input: input{
			manager: getDefaultManager(),
			tasks:   []Task{},
			ueId:    2,
		}}, {
		description: "UeScenarios is nil",
		input: input{
			manager: getDefaultManager(),
			tasks:   nil,
			ueId:    defaultUeId,
		},
	}}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := scenario.input.manager
			ueScenario := UEScenario{
				Tasks: scenario.input.tasks,
			}
			assert.Error(t, s.restartUeScenario(scenario.input.ueId, ueScenario))
		})
	}
}

func TestSetupGnbs(t *testing.T) {
	type input struct {
		manager  ScenarioManager
		gnbConfs []config.GNodeB
		amfConfs []*config.AMF
	}

	type testCases struct {
		description string
		input       input
		expected    int
	}

	gnbConfs1 := []config.GNodeB{
		{
			ControlIF: config.ControlIF{
				Ip:   "127.0.0.1",
				Port: 9487,
			},
			DataIF: config.DataIF{
				Ip:   "127.0.0.2",
				Port: 2152,
			},
			PlmnList: config.PlmnList{
				Mcc:   "999",
				Mnc:   "70",
				Tac:   "000001",
				GnbId: "000001",
			},
			SliceSupportList: config.SliceSupportList{
				Sst: "01",
			},
		},
	}

	gnbConfs2 := []config.GNodeB{
		{
			ControlIF: config.ControlIF{
				Ip:   "192.168.11.12",
				Port: 5000,
			},
			DataIF: config.DataIF{
				Ip:   "192.168.11.12",
				Port: 3000,
			},
			PlmnList: config.PlmnList{
				Mcc:   "999",
				Mnc:   "70",
				Tac:   "000001",
				GnbId: "000002",
			},
			SliceSupportList: config.SliceSupportList{
				Sst: "01",
			},
		},
		{
			ControlIF: config.ControlIF{
				Ip:   "192.168.11.13",
				Port: 9487,
			},
			DataIF: config.DataIF{
				Ip:   "192.168.11.13",
				Port: 2152,
			},
			PlmnList: config.PlmnList{
				Mcc:   "999",
				Mnc:   "70",
				Tac:   "000001",
				GnbId: "000003",
			},
			SliceSupportList: config.SliceSupportList{
				Sst: "01",
			},
		},
	}

	amfConfs := []*config.AMF{{
		Ip:   "192.168.11.30",
		Port: 38412,
	}}

	InitGnb = gnbInitMock

	scenarios := []testCases{{
		description: "1 gnb should be created",
		input: input{
			manager:  getDefaultManager(),
			gnbConfs: gnbConfs1,
			amfConfs: amfConfs,
		},
		expected: len(gnbConfs1),
	}, {
		description: "2 gnbs should be created",
		input: input{
			manager:  getDefaultManager(),
			gnbConfs: gnbConfs2,
			amfConfs: amfConfs,
		},
		expected: len(gnbConfs2),
	},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := scenario.input.manager
			s.setupGnbs(scenario.input.gnbConfs, scenario.input.amfConfs)
			for gnb := range scenario.input.gnbConfs {
				gnbId := scenario.input.gnbConfs[gnb].PlmnList.GnbId
				assert.Equal(t, gnbId, s.gnbs[gnbId].GetGnbId(), "GNB ID between key and values does not match")
			}
			assert.Equal(t, len(scenario.input.gnbConfs), len(s.gnbs), "Configured gnb and created gnb count does not match")
		})
	}
}

func TestSetupUeTaskExecutor(t *testing.T) {
	type input struct {
		manager    ScenarioManager
		ueScenario UEScenario
		ueId       int
		gnbs       *context.GNBContext
	}

	type expected struct {
		ueId   int
		config config.Ue
		gnbs   *context.GNBContext
		tasks  []Task
	}

	type testCases struct {
		description string
		input       input
		expected    expected
	}

	gnbConf := config.GNodeB{
		ControlIF: config.ControlIF{
			Ip:   "127.0.0.1",
			Port: 9487,
		},
		DataIF: config.DataIF{
			Ip:   "127.0.0.2",
			Port: 2152,
		},
		PlmnList: config.PlmnList{
			Mcc:   "999",
			Mnc:   "70",
			Tac:   "000001",
			GnbId: "000001",
		},
		SliceSupportList: config.SliceSupportList{
			Sst: "01",
		},
	}

	amfConfs := []*config.AMF{{
		Ip:   "192.168.11.30",
		Port: 38412,
	}}

	wg := &sync.WaitGroup{}

	gnbs := gnbInitMock(gnbConf, amfConfs, wg)
	tasks := []Task{{
		TaskType: Registration,
		Delay:    2000,
	}, {
		TaskType: NewPDUSession,
		Delay:    0,
	}, {
		TaskType: Idle,
		Delay:    0,
	}}

	scenarios := []testCases{{
		description: "Test normal setup",
		input: input{
			manager: ScenarioManager{
				ues: map[int]*ueTasksCtx{},
			},
			ueScenario: UEScenario{
				Config: config.Ue{
					Msin: "0000001000",
				},
				Tasks: tasks,
			},
			ueId: 1,
			gnbs: gnbs,
		},
		expected: expected{
			ueId: 1,
			config: config.Ue{
				Msin: "0000001000",
			},
			gnbs:  gnbs,
			tasks: tasks,
		},
	},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.description, func(t *testing.T) {
			s := scenario.input.manager
			s.reportChan = make(chan report, 1)
			s.gnbs = map[string]*context.GNBContext{scenario.input.gnbs.GetGnbId(): scenario.input.gnbs}
			s.taskExecutorWg = &sync.WaitGroup{}
			s.setupUeTaskExecutor(scenario.input.ueId, scenario.input.ueScenario)

			time.Sleep(time.Duration(100) * time.Millisecond)

			assert.Equal(t, scenario.expected.tasks, s.ues[scenario.input.ueId].nextTasks)
			assert.Falsef(t, s.ues[scenario.input.ueId].done, "ueTasksCtx should not be done right after initialization")
			select {
			case report := <-s.reportChan:
				assert.Equal(t, scenario.expected.ueId, report.ueId)
				assert.Equal(t, "simulation started", report.reason)
			default:
				t.Error("no response from worker, there might be an issue with channels")
			}
		})
	}

}

func TestSetupUeTaskExecutorHasError(t *testing.T) {

	gnbConf := config.GNodeB{
		ControlIF: config.ControlIF{
			Ip:   "127.0.0.1",
			Port: 9487,
		},
		DataIF: config.DataIF{
			Ip:   "127.0.0.2",
			Port: 2152,
		},
		PlmnList: config.PlmnList{
			Mcc:   "999",
			Mnc:   "70",
			Tac:   "000001",
			GnbId: "000001",
		},
		SliceSupportList: config.SliceSupportList{
			Sst: "01",
		},
	}

	amfConfs := []*config.AMF{{
		Ip:   "192.168.11.30",
		Port: 38412,
	}}

	wg := &sync.WaitGroup{}

	gnb := gnbInitMock(gnbConf, amfConfs, wg)
	tasks := []Task{{
		TaskType: Registration,
		Delay:    2000,
	}, {
		TaskType: NewPDUSession,
		Delay:    0,
	}, {
		TaskType: Idle,
		Delay:    0,
	}}
	ueScenario := UEScenario{
		Config: config.Ue{
			Msin: "0000001000",
		},
		Tasks: tasks,
	}

	t.Run("invalid UE Id", func(t *testing.T) {

		s := getDefaultManager()
		s.reportChan = make(chan report, 1)
		s.gnbs = map[string]*context.GNBContext{gnb.GetGnbId(): gnb}
		s.taskExecutorWg = &sync.WaitGroup{}
		ueid := 0
		assert.Error(t, s.setupUeTaskExecutor(ueid, ueScenario))
	})
	t.Run("Task Executor already exists", func(t *testing.T) {

		s := getDefaultManager()
		s.reportChan = make(chan report, 1)
		s.gnbs = map[string]*context.GNBContext{gnb.GetGnbId(): gnb}
		s.taskExecutorWg = &sync.WaitGroup{}
		ueid := 0
		s.setupUeTaskExecutor(ueid, ueScenario)
		assert.Error(t, s.setupUeTaskExecutor(ueid, ueScenario))
	})
}

func TestStartOrderRateController(t *testing.T) {
	orderCount := 20
	timeBetweenRegistration := 500
	orderChan := make(chan order, 5)
	stopChan := make(chan int, 1)
	taskChan := make(chan Task, 1)
	go startOrderRateController(orderChan, timeBetweenRegistration)

	go func() {
		for i := 0; i < orderCount; i++ {
			orderChan <- order{task: Task{}, workerChan: taskChan}
		}
	}()
	time.Sleep(time.Duration(timeBetweenRegistration/2) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(timeBetweenRegistration) * time.Millisecond)
	endTicker := time.NewTicker(time.Duration((orderCount+5)*timeBetweenRegistration) * time.Millisecond)
	expected := 0
	actual := 0
	stop := false
	for !stop {
		select {
		case <-taskChan:
			actual++
		case <-ticker.C:
			if expected < orderCount {
				expected++
				assert.Equal(t, expected, actual)
			}
		case <-endTicker.C:
			stop = true
			stopChan <- 0
		}
	}
	close(orderChan)
	assert.Equal(t, orderCount, actual)
}

func TestCleanup(t *testing.T) {

	s := getDefaultManager()
	s.taskExecutorWg.Add(1)
	time.AfterFunc(time.Duration(1)*time.Second, func() { s.taskExecutorWg.Done() })
	s.reportChan = make(chan report, 2)
	s.reportChan <- report{ueId: 1, reason: "test"}
	s.reportChan <- report{ueId: 2, reason: "test"}

	s.registrationQueue <- order{}

	s.cleanup()

	select {
	case task := <-s.ues[1].workerChan:
		assert.Equal(t, Terminate, task.TaskType)
	default:
		t.Error("Terminating task not send to UE")
	}

	_, open := <-s.reportChan
	assert.False(t, open, "report channel should be closed")
	_, open = <-s.registrationQueue
	assert.False(t, open, "registrationQueue channel should be closed")
}
