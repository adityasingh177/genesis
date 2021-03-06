/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package controller

import (
	"fmt"
	"testing"
	"time"

	queue "github.com/whiteblock/genesis/mocks/amqp"
	handler "github.com/whiteblock/genesis/mocks/pkg/handler"
	"github.com/whiteblock/genesis/pkg/entity"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCommandController_Failure(t *testing.T) {
	ctl, err := NewCommandController(0, nil, nil, nil, nil, nil, logrus.New())
	assert.Nil(t, ctl)
	assert.Error(t, err)
}

func TestNewCommandController_Ignore_CreateQueueFailure(t *testing.T) {
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv3 := new(queue.AMQPService)
	serv4 := new(queue.AMQPService)
	serv.On("CreateQueue").Return(fmt.Errorf("err")).Once()
	serv2.On("CreateQueue").Return(fmt.Errorf("err")).Once()
	serv3.On("CreateQueue").Return(fmt.Errorf("err")).Once()
	serv4.On("CreateQueue").Return(fmt.Errorf("err")).Once()

	control, err := NewCommandController(2, serv, serv3, serv2, serv4, nil, logrus.New())
	assert.NotNil(t, control)
	assert.NoError(t, err)

	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
	serv3.AssertExpectations(t)
}

func TestCommandController_Consumption(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv3 := new(queue.AMQPService)
	serv4 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()
	serv3.On("CreateQueue").Return(nil).Once()
	serv4.On("CreateQueue").Return(nil).Once()
	serv4.On("Send", mock.Anything).Return(nil)
	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Run(func(_ mock.Arguments) {
		processedChan <- true
	}).Return(amqp.Publishing{}, amqp.Publishing{}, entity.NewSuccessResult()).Times(items)

	control, err := NewCommandController(2, serv, serv3, serv2, serv4, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
}

func TestCommandController_ConsumptionAllDone(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv3 := new(queue.AMQPService)
	serv4 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()
	serv3.On("CreateQueue").Return(nil).Once()
	serv4.On("CreateQueue").Return(nil).Once()
	serv4.On("Send", mock.Anything).Return(nil)
	serv3.On("Send", mock.Anything).Return(nil)
	serv2.On("Send", mock.Anything).Return(nil).Times(items).Run(func(_ mock.Arguments) {
		processedChan <- true
	})

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{}, amqp.Publishing{},
		entity.NewAllDoneResult()).Times(items)

	control, err := NewCommandController(2, serv, serv3, serv2, serv4, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
}

func TestCommandController_ConsumptionAllDone_Send_Err(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv2 := new(queue.AMQPService)
	serv3 := new(queue.AMQPService)
	serv4 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv2.On("CreateQueue").Return(nil).Once()
	serv2.On("Send", mock.Anything).Return(fmt.Errorf("err")).Times(items).Run(func(_ mock.Arguments) {
		processedChan <- true
	})
	serv4.On("Send", mock.Anything).Return(nil)
	serv3.On("CreateQueue").Return(nil).Once()
	serv4.On("CreateQueue").Return(nil).Once()

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{}, amqp.Publishing{},
		entity.NewAllDoneResult()).Times(items)

	control, err := NewCommandController(2, serv, serv3, serv2, serv4, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
	serv2.AssertExpectations(t)
}

func TestCommandController_Requeue(t *testing.T) {
	items := 10

	processedChan := make(chan bool, items)
	deliveryChan := make(chan amqp.Delivery, items)
	serv := new(queue.AMQPService)
	serv3 := new(queue.AMQPService)
	serv4 := new(queue.AMQPService)
	serv.On("Consume").Return((<-chan amqp.Delivery)(deliveryChan), nil).Once()
	serv.On("CreateQueue").Return(nil).Once()
	serv.On("Requeue", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		processedChan <- true
	}).Return(nil).Times(items)
	serv2 := new(queue.AMQPService)
	serv2.On("CreateQueue").Return(nil).Once()
	serv3.On("CreateQueue").Return(nil).Once()
	serv4.On("CreateQueue").Return(nil).Once()
	serv4.On("Send", mock.Anything).Return(nil)

	hand := new(handler.DeliveryHandler)
	hand.On("Process", mock.Anything).Return(amqp.Publishing{}, amqp.Publishing{},
		entity.NewErrorResult(fmt.Errorf("some non-fatal error"))).Times(items)

	control, err := NewCommandController(2, serv, serv3, serv2, serv4, hand, logrus.New())
	assert.Equal(t, err, nil)
	go control.Start()

	for i := 0; i < items; i++ {
		deliveryChan <- amqp.Delivery{}
	}
	for i := 0; i < items; i++ {
		select {
		case <-processedChan:
		case <-time.After(5 * time.Second):
			t.Fatal("messages were not consumed within 5 seconds")
		}
	}

	close(deliveryChan)
	hand.AssertExpectations(t)
	serv.AssertExpectations(t)
}
