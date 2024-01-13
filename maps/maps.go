package maps

import (
	"github.com/dolthub/maphash"
	"math/bits"
)

const bucketSize = 8

type Map struct {
	Length          int
	LogBucketsCount uint8 // 2-логарифм числа бакетов
	Buckets         []Bucket
	hasher          maphash.Hasher[string]
}

type Bucket struct {
	tophash  [bucketSize]uint8  // массив первых восьми бит от хэша ключа
	keys     [bucketSize]string // массив ключей
	values   [bucketSize]string // массив значений
	overflow *Bucket            // ссылка на бакет в случае переполнения текущего
}

func New(size uint16) Map {
	// из кол-ва элементов получаем необходимое число бакетов
	var requiredBucketsNum = size>>3 + 1
	// храним не кол-во, а логарифм по 2
	log2 := bits.Len16(requiredBucketsNum)
	// считаем кол-во бакетов, которые нужно создать
	bucketNum := 1 << log2
	return Map{
		LogBucketsCount: uint8(log2),
		Buckets:         make([]Bucket, bucketNum),
		hasher:          maphash.NewHasher[string](),
	}
}

func (m *Map) Add(key, value string) {
	keyHash := uint8(m.hasher.Hash(key))
	// вычисляем id бакета (low order bits)
	bucketID := keyHash & m.LogBucketsCount
	// вычисляем топ хэш ключа для быстрого поиска (high order bits)
	tophash := keyHash >> 7
	isNew := m.Buckets[bucketID].put(key, value, tophash)
	// если добавлен новый ключ - инкремент длины мапы
	if isNew {
		m.Length++
	}
	return
}

func (m *Map) Get(key string) (string, bool) {
	keyHash := uint8(m.hasher.Hash(key))
	// вычисляем id бакета (low order bits)
	bucketID := keyHash & m.LogBucketsCount
	// вычисляем топ хэш ключа для быстрого поиска (high order bits)
	tophash := keyHash >> 7
	return m.Buckets[bucketID].get(key, tophash)
}

func (m *Map) Delete(key string) {
	keyHash := m.hasher.Hash(key)
	_ = keyHash
}

func (b *Bucket) put(key, value string, tophash uint8) bool {
	var emptyIdx *int
	// ищем ключ по старшим битам хэша
	for i := range b.tophash {
		if b.tophash[i] != tophash {
			// если хэш не совпал и слот свободен - сохраняем id слота (будем писать в него значение, если ключ новый)
			if b.tophash[i] == 0 && emptyIdx == nil {
				emptyIdx = new(int)
				*emptyIdx = i
			}
			continue
		}

		// дополнительно сверяем сам ключ
		if b.keys[i] != key {
			continue
		}

		// пишем новое значение
		b.values[i] = value
		// возвращаем false, т.к. не был занят новый слот
		return false
	}

	// не нашли в бакете, проверяем overflow бакет
	// если нет пустых слотов или уже было переполнение (создан overflow бакет)
	if emptyIdx == nil || b.overflow != nil {
		// создаем overflow бакет, если его не было
		if b.overflow == nil {
			b.overflow = &Bucket{}
		}
		return b.overflow.put(key, value, tophash)
	}

	// не нашли ключ в overflow бакете - ключ новый - пишем в свободный слот бакета
	b.keys[*emptyIdx] = key
	b.values[*emptyIdx] = value
	b.tophash[*emptyIdx] = tophash
	// возвращаем true, т.к. занят новый слот
	return true
}

func (b *Bucket) get(key string, tophash uint8) (string, bool) {
	for i := 0; i < bucketSize; i++ {
		if b.tophash[i] == tophash {
			if b.keys[i] == key {
				return b.values[i], true
			}
		}
	}
	if b.overflow != nil {
		return b.overflow.get(key, tophash)
	}
	return "", false
}
