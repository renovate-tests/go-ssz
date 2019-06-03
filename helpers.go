package ssz

import (
	"math"
)

// Given ordered objects of the same basic type, serialize them, pack them into BYTES_PER_CHUNK-byte
// chunks, right-pad the last chunk with zero bytes, and return the chunks.
// Basic types are either bool, or uintN where N = {8, 16, 32, 64, 128, 256}.
//
// Important: due to limitations in Go generics, we will assume the input is already
// a list of SSZ-encoded objects of the same type.
func pack(serializedItems [][]byte) ([][]byte, error) {
	// If there are no items, we return an empty chunk.
	if len(serializedItems) == 0 {
		emptyChunk := make([]byte, BytesPerChunk)
		return [][]byte{emptyChunk}, nil
	// If each item has exactly BYTES_PER_CHUNK length, we return the list of serialized items.
	} else if len(serializedItems[0]) == BytesPerChunk {
		return serializedItems, nil
	}
	// We flatten the list in order to pack its items into byte chunks correctly.
	orderedItems := []byte{}
	for _, item := range serializedItems {
		orderedItems = append(orderedItems, item...)
	}
	numItems := len(orderedItems)
	chunks := [][]byte{}
	for i := 0; i < numItems; i += BytesPerChunk {
		j := i + BytesPerChunk
		// We create our upper bound index of the chunk, if it is greater than numItems,
		// we set it as numItems itself.
		if j > numItems {
			j = numItems
		}
		// We create chunks from the list of serialized items based on the
		// indices determined above.
		chunks = append(chunks, orderedItems[i:j])
	}
	// Right-pad the last chunk with zero bytes if it does not
	// have length BytesPerChunk.
	lastChunk := chunks[len(chunks)-1]
	for len(lastChunk) < BytesPerChunk {
		lastChunk = append(lastChunk, 0)
	}
	chunks[len(chunks)-1] = lastChunk
	return chunks, nil
}

// Given ordered BYTES_PER_CHUNK-byte chunks, if necessary append zero chunks so that the
// number of chunks is a power of two, Merkleize the chunks, and return the root.
// Note that merkleize on a single chunk is simply that chunk, i.e. the identity
// when the number of chunks is one.
func merkleize(chunks [][]byte) ([32]byte, error) {
	if len(chunks) == 1 {
		var root [32]byte
		copy(root[:], chunks[0])
		return root, nil
	}
	for !isPowerTwo(len(chunks)) {
		chunks = append(chunks, make([]byte, BytesPerChunk))
	}
	hashLayer := make([][32]byte, len(chunks))
	// We keep track of the hash layers of a Merkle trie until we reach
	// the top layer of length 1, which contains the single root element.
	//        [Root]      -> Top layer has length 1.
	//    [E]       [F]   -> This layer has length 2.
	// [A]  [B]  [C]  [D] -> The bottom layer has length 4 (needs to be a power of two.
	for len(hashLayer) > 1 {
		layer := [][32]byte{}
		for i := 0; i < len(hashLayer); i += 2 {
			hashedChunk := Hash(append(chunks[i], chunks[i+1]...))
			layer = append(layer, hashedChunk)
		}
		hashLayer = layer
	}
	return hashLayer[0], nil
}

// Given a Merkle root root and a length length ("uint256" little-endian serialization)
// return hash(root + length).
func mixInLength(root [32]byte, length []byte) [32]byte {
	return Hash(append(root[:], length...))
}

// Given a Merkle root root and a type_index type_index ("uint256" little-endian serialization)
// return hash(root + type_index).
func mixInType(root [32]byte, typeIndex []byte) [32]byte {
	return Hash(append(root[:], typeIndex...))
}

// fast verification to check if an number if a power of two.
func isPowerTwo(num int) bool {
	// log base 2 of 0 is undefined, so we return false.
	if num == 0 {
		return false
	}
	elem := math.Log2(float64(num))
	return math.Floor(elem) == math.Ceil(elem)
}
