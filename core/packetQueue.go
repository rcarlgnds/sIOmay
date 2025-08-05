package core

import (
	"errors"
)


//===================IMPORTANT==================
// this code does not include mutex in order to gain max performance
// ONLY USED IT IN SINGLE PRODUCER-CONSUMER PATTERN
//==============================================
type PacketQueue struct {
	buffer [][]byte
	maxSize int16
	head int16
	tail int16
}

func NewPacketQueue(size int16) *PacketQueue {
	return &PacketQueue{
		buffer:  make([][]byte, size),
		maxSize: size,
	}
}

func (q *PacketQueue) Append(packet []byte) error{
	if (q.tail+1)%q.maxSize == q.head  {
		return errors.New("Queue is full")
	}
	q.buffer[q.tail] = packet;
	q.tail += 1
	q.tail %= q.maxSize;
	return nil;
}

func (q *PacketQueue) Pop() ([]byte , error){
	if q.head == q.tail {
		return nil, errors.New("Queue is empty")
	}
	data  := q.buffer[q.head]
	q.head +=1
	q.head %= q.maxSize
	return data, nil
}

func (q *PacketQueue) Length() int{
	
	if q.tail >= q.head{
		return int(q.tail - q.head)
	}else{
		return int(q.maxSize-q.head+q.tail)
	}
}