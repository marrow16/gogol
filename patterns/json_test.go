package patterns

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPattern_MarshalJSON(t *testing.T) {
	cells := []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false}
	p, err := NewPattern("Glider", 5, cells)
	require.NoError(t, err)

	data, err := json.Marshal(p)
	require.NoError(t, err)
	require.Equal(t, `{"name":"Glider","width":5,"height":5,"rule":"B3/S23","cells":{"rle":"$3bo$bobo$2b2o"}}`, string(data))

	var p2 Pattern
	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)
	require.Equal(t, "Glider", p2.Name)
	require.Equal(t, 5, p2.Width)
	require.Equal(t, 5, p2.Height)
	require.Equal(t, cells, p2.Cells)

	pe := Pattern{}
	_, err = json.Marshal(pe)
	require.Error(t, err)

	p.CellsEncoding = Base64CellsEncoding
	data, err = json.Marshal(p)
	require.NoError(t, err)
	require.Equal(t, `{"name":"Glider","width":5,"height":5,"rule":"B3/S23","cells":{"base64":"AJRgAA=="}}`, string(data))

	err = json.Unmarshal(data, &p2)
	require.NoError(t, err)
	require.Equal(t, "Glider", p2.Name)
	require.Equal(t, 5, p2.Width)
	require.Equal(t, 5, p2.Height)
	require.Equal(t, cells, p2.Cells)
}

func TestPattern_UnmarshalJSON_Errors(t *testing.T) {
	testCases := []struct {
		data string
		err  string
	}{
		{
			data: `invalid json`,
			err:  "invalid character",
		},
		{
			data: `{"width":-1}`,
			err:  "cannot unmarshal number",
		},
		{
			data: `{"height":-1}`,
			err:  "cannot unmarshal number",
		},
		{
			data: `{"rule":"invalid"}`,
			err:  "invalid RLE rule",
		},
		{
			data: `{}`, // no cells data
			err:  "no cells defined",
		},
		{
			data: `{"cells": {"rle":"bad"}}`,
			err:  "invalid RLE format",
		},
		{
			data: `{"cells": {"base64":"x"}}`,
			err:  "illegal base64 data",
		},
		{
			data: `{"cells": {"base64":"AJRgAA=="}}`,
			err:  "invalid cell data length",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			p := &Pattern{}
			err := p.UnmarshalJSON([]byte(tc.data))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err)
		})
	}
}

func TestPackCells(t *testing.T) {
	cells := []bool{
		false, false, false, false, false, false, false, false,
		true, false, false, true, false, true, false, false,
		false, true, true, false, false, false, false, false,
		false}
	data := PackCells(cells)
	require.Equal(t, []byte{
		0b00000000,
		0b10010100,
		0b01100000,
		0b00000000}, data)

	cells = []bool{
		false, false, true,
		true, false, true,
		false, true, true}
	data = PackCells(cells)
	require.Equal(t, []byte{
		0b00110101,
		0b10000000}, data)
}

func TestUnpackCells(t *testing.T) {
	cells, err := UnpackCells([]byte{
		0b00000000,
		0b10010100,
		0b01100000,
		0b00000000}, 5, 5)
	require.NoError(t, err)
	require.Equal(t, []bool{
		false, false, false, false, false,
		false, false, false, true, false,
		false, true, false, true, false,
		false, false, true, true, false,
		false, false, false, false, false}, cells)

	cells, err = UnpackCells([]byte{
		0b00110101,
		0b10000000}, 3, 3)
	require.NoError(t, err)
	require.Equal(t, []bool{
		false, false, true,
		true, false, true,
		false, true, true}, cells)
}

func TestUnpackCells_Errors(t *testing.T) {
	t.Run("not enough bytes", func(t *testing.T) {
		_, err := UnpackCells([]byte{}, 3, 3)
		require.Error(t, err)
	})
	t.Run("too many bytes", func(t *testing.T) {
		_, err := UnpackCells([]byte{0x00, 0x00}, 2, 2)
		require.Error(t, err)
	})
	t.Run("non-zero padding bits", func(t *testing.T) {
		_, err := UnpackCells([]byte{0x01}, 2, 2)
		require.Error(t, err)
	})
}

func TestRleCells(t *testing.T) {
	testCases := []struct {
		width     int
		height    int
		cells     []bool
		expect    string
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			width:     -1,
			height:    1,
			expectErr: true,
		},
		{
			width:     1,
			height:    -1,
			expectErr: true,
		},
		{
			width:  5,
			height: 5,
			cells: []bool{
				false, false, false, false, false,
				false, false, false, true, false,
				false, true, false, true, false,
				false, false, true, true, false,
				false, false, false, false, false},
			expect: "$3bo$bobo$2b2o",
		},
		{
			width:  5,
			height: 5,
			cells: []bool{
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false},
			expect: "",
		},
		{
			width:  5,
			height: 5,
			cells: []bool{
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, true},
			expect: "4$4bo",
		},
		{
			width:  5,
			height: 5,
			cells: []bool{
				true, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false,
				false, false, false, false, false},
			expect: "o",
		},
		{
			width:     5,
			height:    5,
			cells:     []bool{},
			expectErr: true,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			s, err := RleCells(tc.width, tc.height, tc.cells)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expect, s)
				cells, err := ParseRLEData(tc.width, tc.height, s)
				require.NoError(t, err)
				assert.Equal(t, tc.cells, cells)
			}
		})
	}
}
