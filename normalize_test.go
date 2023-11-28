package cadeft

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	sampleText := `Álfréd, the hândŷ mân, häd ēârnėd a rěsćūě. He dîd not ùsě cômplètě cîrcûmfle chèmîcáls, âlthoûgh hě hăd an ōppôrtûnĭtý. His jôb wäs to mĕasure ćērtăin tȇmpérâtûrēs, whîle bėâring a lĭttle mâcron êmblēm. He nėver sǫught to âssûme thě sȗbtlě dîaresîs, prėferrĭng to cõncěntrāte on cédilla mèâsureměnts ând ťìlděd äccentūâtions. Hė wãs rěgǫrděd as thě môst skīllful of thě sènsiblě sųmmed ǫgněk ênthusīâsts.`
	expected := `Alfred, the handy man, had earned a rescue. He did not use complete circumfle chemicals, although he had an opportunity. His job was to measure certain temperatures, while bearing a little macron emblem. He never sought to assume the subtle diaresis, preferring to concentrate on cedilla measurements and tilded accentuations. He was regorded as the most skillful of the sensible summed ognek enthusiasts.`
	normal, err := normalize(sampleText)
	require.NoError(t, err)
	require.Equal(t, expected, normal)

	// We attempt to normalize diacritical but non ascii and non diacritical letters are errors.
	notAscii := `Лорем ипсум Λορεμ ι成空援光必 मजबुत प्रमान वर्तमान`
	_, err = normalize(notAscii)
	require.Error(t, err)
}
