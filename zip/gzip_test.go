package zip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GZip(t *testing.T) {
	text := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras ullamcorper urna purus, vitae varius sem pellentesque sed. Etiam elementum nisl sed dolor tincidunt, vitae tristique mauris rutrum. In quam magna, ultricies id eleifend quis, volutpat eu velit. Suspendisse non aliquam dolor. Nunc consequat in nunc a efficitur. Curabitur sed lectus lorem. Aliquam accumsan id ipsum eget euismod. Aliquam et nulla neque. In sit amet mi nec metus facilisis aliquet. Nulla sodales, ex vitae mollis tempus, ante nulla porta dolor, sed elementum eros enim in magna. Nullam vel massa nisl."
	compressed, err := GZipData([]byte(text))
	assert.Nil(t, err)
	text2, err := GUnzipData(compressed)
	assert.Nil(t, err)
	assert.Equal(t, string(text2), text)
}
