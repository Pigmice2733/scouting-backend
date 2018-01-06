// Code generated by go-bindata.
// sources:
// 0_create_users_table.up.sql
// 0_drop_users_table.down.sql
// 1_create_events_table.up.sql
// 1_drop_events_table.down.sql
// 2_create_matches_table.up.sql
// 2_drop_matches_table.down.sql
// 3_create_alliances_table.up.sql
// 3_drop_alliances_table.down.sql
// 4_create_reports_table.up.sql
// 4_drop_reports_table.down.sql
// 5_add_event_type.up.sql
// 5_remove_event_type.down.sql
// bindata.go
// DO NOT EDIT!

package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var __0_create_users_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x0e\x72\x75\x0c\x71\x55\x08\x71\x74\xf2\x71\x55\xf0\x74\x53\xf0\xf3\x0f\x51\x70\x8d\xf0\x0c\x0e\x09\x56\x28\x2d\x4e\x2d\x2a\x56\xd0\xe0\xe2\x04\x31\xf2\x12\x73\x53\x15\x42\x5c\x23\x42\xc0\x2a\xfc\x42\x7d\x7c\x14\x42\xfd\x3c\x03\x43\x5d\x75\xb8\x38\x33\x12\x8b\x33\x52\x53\x02\x12\x8b\x8b\xcb\xf3\x8b\x52\x50\x55\x71\x69\x02\x02\x00\x00\xff\xff\xe1\x96\xd7\xd2\x62\x00\x00\x00")

func _0_create_users_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__0_create_users_tableUpSql,
		"0_create_users_table.up.sql",
	)
}

func _0_create_users_tableUpSql() (*asset, error) {
	bytes, err := _0_create_users_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "0_create_users_table.up.sql", size: 98, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __0_drop_users_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x28\x2d\x4e\x2d\x2a\xb6\x06\x04\x00\x00\xff\xff\xbd\x6d\xc5\x8d\x11\x00\x00\x00")

func _0_drop_users_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__0_drop_users_tableDownSql,
		"0_drop_users_table.down.sql",
	)
}

func _0_drop_users_tableDownSql() (*asset, error) {
	bytes, err := _0_drop_users_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "0_drop_users_table.down.sql", size: 17, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_create_events_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x3c\x8b\x4d\x0a\xc2\x30\x10\x85\xd7\x99\x53\xbc\xa5\x42\x2f\x11\x65\x84\x60\x12\x4b\xf2\x84\xd6\x5d\xc1\x01\x41\xac\x60\x83\xe0\xed\x45\x44\xb7\xdf\xcf\xb6\xa8\xa7\x82\x7e\x13\x15\x61\x87\x7c\x20\x74\x08\x95\x15\xf6\xb4\xb9\x2d\x58\x89\xbb\xda\x0b\xd4\x81\xe8\x4b\x48\xbe\x8c\xd8\xeb\xd8\x89\x9b\xa7\x9b\x7d\xf9\xe7\xca\xc7\x18\x3b\x71\xcb\xe5\xfe\x68\xf9\x67\x3a\x71\xe7\xa9\x19\x18\x92\x56\xfa\xd4\xf3\xf4\x8f\x65\xfd\x0e\x00\x00\xff\xff\x50\xda\x81\x7d\x7d\x00\x00\x00")

func _1_create_events_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_create_events_tableUpSql,
		"1_create_events_table.up.sql",
	)
}

func _1_create_events_tableUpSql() (*asset, error) {
	bytes, err := _1_create_events_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_create_events_table.up.sql", size: 125, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __1_drop_events_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\x2d\x4b\xcd\x2b\x29\x06\x04\x00\x00\xff\xff\x27\xe5\x89\x64\x11\x00\x00\x00")

func _1_drop_events_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__1_drop_events_tableDownSql,
		"1_drop_events_table.down.sql",
	)
}

func _1_drop_events_tableDownSql() (*asset, error) {
	bytes, err := _1_drop_events_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "1_drop_events_table.down.sql", size: 17, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __2_create_matches_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8e\x4d\x6a\xc3\x30\x10\x46\xd7\xd2\x29\x66\x69\x83\x2e\xa1\x84\x71\x10\x91\xa5\x20\x4d\x69\xd2\x9d\x2b\x0f\x34\xf8\xaf\xd8\x72\xc1\xb7\x2f\x2e\xf5\xa6\x74\xfb\xde\xf7\xc1\x3b\x07\xd4\x84\x40\xfa\x64\x11\x4c\x05\xce\x13\xe0\xdd\x44\x8a\x30\x34\x39\x7d\xf0\x02\x85\x14\x1d\x6f\x40\x78\x27\xb8\x05\x53\xeb\xf0\x80\x2b\x3e\x94\x14\xfc\xc5\x63\xbe\x1e\x6e\xbf\xba\x17\x6b\x95\x14\x9f\x33\xb7\xcf\x94\xb9\xa5\xe7\xc0\x40\xa6\xc6\x48\xba\xbe\xd1\x9b\x92\xa2\x49\x79\x6d\xfa\x7f\xc4\x7b\xbf\xf2\xeb\x34\xc2\xc9\x7b\x8b\xda\x29\x29\x66\x6e\x63\x9a\x66\x06\xe3\x08\x2f\x18\x7e\x47\x7f\x59\xe5\x03\x9a\x8b\xdb\xa3\x8a\x23\xa9\x84\x80\x15\x06\x74\x67\x8c\xf0\x03\x97\xa2\xe3\xad\x94\xe5\x77\x00\x00\x00\xff\xff\x95\xc6\xa5\x34\xf2\x00\x00\x00")

func _2_create_matches_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__2_create_matches_tableUpSql,
		"2_create_matches_table.up.sql",
	)
}

func _2_create_matches_tableUpSql() (*asset, error) {
	bytes, err := _2_create_matches_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "2_create_matches_table.up.sql", size: 242, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __2_drop_matches_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\xc8\x4d\x2c\x49\xce\x48\x2d\x06\x04\x00\x00\xff\xff\xed\x06\x12\x35\x12\x00\x00\x00")

func _2_drop_matches_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__2_drop_matches_tableDownSql,
		"2_drop_matches_table.down.sql",
	)
}

func _2_drop_matches_tableDownSql() (*asset, error) {
	bytes, err := _2_drop_matches_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "2_drop_matches_table.down.sql", size: 18, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __3_create_alliances_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x8c\x41\x0a\x83\x30\x14\x44\xd7\xe6\x14\xb3\x34\xe0\x25\x54\xc6\x12\x0c\x09\x8d\x11\xec\xd2\xca\x87\x4a\xd5\x45\xad\x0b\x6f\x5f\xb0\xb4\x14\xba\x9d\x79\xef\x95\x81\x79\x24\x62\x5e\x58\xc2\x54\x70\x3e\x82\x9d\x69\x62\x83\x7e\x9a\xc6\x7e\x19\x64\x45\xaa\x92\xb9\x7f\x0e\xb7\x5a\x76\x44\x76\xf1\xa0\x5c\x6b\x6d\xa6\x92\x71\x2d\xa6\x4d\x50\x78\x6f\x99\xbb\xdf\x67\xd9\xe6\xab\x3c\xfe\x84\xca\x07\x9a\x93\x43\xcd\x4b\xfa\xa9\x6a\x04\x56\x0c\x74\x25\x1b\x1c\xa3\xac\xe9\x5d\x76\x9d\xa9\xa4\x75\xe6\xdc\xf2\x8b\x66\x78\x77\xb5\xd2\xaf\x00\x00\x00\xff\xff\x80\x51\xbd\xfc\xbc\x00\x00\x00")

func _3_create_alliances_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__3_create_alliances_tableUpSql,
		"3_create_alliances_table.up.sql",
	)
}

func _3_create_alliances_tableUpSql() (*asset, error) {
	bytes, err := _3_create_alliances_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "3_create_alliances_table.up.sql", size: 188, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __3_drop_alliances_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x48\xcc\xc9\xc9\x4c\xcc\x4b\x4e\x2d\x06\x04\x00\x00\xff\xff\x17\xe1\x40\xdc\x14\x00\x00\x00")

func _3_drop_alliances_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__3_drop_alliances_tableDownSql,
		"3_drop_alliances_table.down.sql",
	)
}

func _3_drop_alliances_tableDownSql() (*asset, error) {
	bytes, err := _3_drop_alliances_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "3_drop_alliances_table.down.sql", size: 20, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __4_create_reports_tableUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8e\xc1\x6a\x86\x30\x10\x84\xcf\xe6\x29\xf6\x98\x40\x5e\x42\x7f\xd6\x12\x0c\x09\x8d\x11\xec\x31\x94\x85\x96\x56\x5b\x92\x58\xf0\xed\x8b\x4a\x40\x6b\x2f\x21\xcc\x7c\x3b\x33\x0f\x87\xb5\x47\xf0\x75\xa3\x11\x54\x0b\xc6\x7a\xc0\x51\xf5\xbe\x87\x48\xdf\x5f\x31\x27\xe0\xac\x3a\xbe\x14\xc1\xe3\xe8\x77\xc6\x0c\x5a\x4b\x56\xd1\x0f\xcd\xb9\xa3\xf5\x66\x4c\x21\xbf\xbe\xfd\x67\xbc\xa7\xe6\x73\x21\x68\xac\xd5\x58\x9b\xb3\x93\x29\x4c\x37\x3c\xe5\x90\xd3\x4d\x1d\x8c\x7a\x1e\x90\x97\x76\x09\xa5\x4e\xc2\x96\x22\x24\x03\x68\xad\x43\xf5\x64\xa0\xc3\x17\x5e\xf6\x0b\x70\xd8\xa2\x43\xf3\xc0\x1e\x96\x44\x31\xf1\xed\x9d\xc3\x44\x42\xb2\xea\x7c\x52\xb2\x2f\x27\xbb\x98\xf8\x07\xad\x7f\xf1\x32\xe0\x82\xef\x22\x1d\x3c\x13\xbf\x01\x00\x00\xff\xff\xf1\xac\xe6\x8d\x6b\x01\x00\x00")

func _4_create_reports_tableUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__4_create_reports_tableUpSql,
		"4_create_reports_table.up.sql",
	)
}

func _4_create_reports_tableUpSql() (*asset, error) {
	bytes, err := _4_create_reports_tableUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "4_create_reports_table.up.sql", size: 363, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __4_drop_reports_tableDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x09\xf2\x0f\x50\x08\x71\x74\xf2\x71\x55\x28\x4a\x2d\xc8\x2f\x2a\x29\xb6\x06\x04\x00\x00\xff\xff\x55\x5c\x72\x90\x13\x00\x00\x00")

func _4_drop_reports_tableDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__4_drop_reports_tableDownSql,
		"4_drop_reports_table.down.sql",
	)
}

func _4_drop_reports_tableDownSql() (*asset, error) {
	bytes, err := _4_drop_reports_tableDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "4_drop_reports_table.down.sql", size: 19, mode: os.FileMode(436), modTime: time.Unix(1514925798, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __5_add_event_typeUpSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x48\x2d\x4b\xcd\x2b\x29\x56\x70\x74\x71\x51\x70\xf6\xf7\x09\xf5\xf5\x83\x88\x84\x54\x16\xa4\x2a\x78\xfa\x85\xb8\xba\xbb\x06\x29\xf8\xf9\x87\x28\xf8\x85\xfa\xf8\x00\x02\x00\x00\xff\xff\x47\xcd\xa1\xbe\x38\x00\x00\x00")

func _5_add_event_typeUpSqlBytes() ([]byte, error) {
	return bindataRead(
		__5_add_event_typeUpSql,
		"5_add_event_type.up.sql",
	)
}

func _5_add_event_typeUpSql() (*asset, error) {
	bytes, err := _5_add_event_typeUpSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "5_add_event_type.up.sql", size: 56, mode: os.FileMode(436), modTime: time.Unix(1515135154, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var __5_remove_event_typeDownSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\xf4\x09\x71\x0d\x52\x08\x71\x74\xf2\x71\x55\x48\x2d\x4b\xcd\x2b\x29\x56\x70\x09\xf2\x0f\x50\x70\xf6\xf7\x09\xf5\xf5\x83\x08\x85\x54\x16\xa4\x02\x02\x00\x00\xff\xff\xfa\x1d\x58\x8e\x28\x00\x00\x00")

func _5_remove_event_typeDownSqlBytes() ([]byte, error) {
	return bindataRead(
		__5_remove_event_typeDownSql,
		"5_remove_event_type.down.sql",
	)
}

func _5_remove_event_typeDownSql() (*asset, error) {
	bytes, err := _5_remove_event_typeDownSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "5_remove_event_type.down.sql", size: 40, mode: os.FileMode(436), modTime: time.Unix(1515135161, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _bindataGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x5a\xdf\x6f\xdb\x38\x12\x7e\xb6\xfe\x0a\x6d\x80\x5d\xd8\x87\x5c\x22\xea\xb7\x02\xf4\x65\xdb\x1e\xd0\x87\xeb\x02\xb7\xdd\xa7\xe3\x21\xa0\x28\x32\x2b\x9c\x6d\xa5\xb2\xdd\xb2\x2d\xf2\xbf\x1f\x3e\x0e\x95\x38\xb6\x24\x7b\x5d\xa3\xb9\x07\xd9\xfa\xc5\xe1\x0c\x39\xf3\x7d\x33\xa4\xae\xaf\xfd\xd7\x4d\xa5\xfc\x3b\xb5\x54\xad\x58\xab\xca\x2f\xbf\xf8\x77\xcd\xdf\xcb\x7a\x59\x89\xb5\xb8\xf2\xae\xaf\xfd\x55\xb3\x69\xa5\x5a\xdd\xe0\x3c\xb8\x95\xad\x12\x6b\x75\xbb\x59\xa9\x76\x75\xbb\x16\xe5\x5c\x5d\x6d\xee\xaf\x56\x1f\xe7\xf4\xb8\x6a\x9b\xfb\x67\x0f\xab\xe6\xf3\xb2\x7b\xcc\xba\xd6\xea\x93\x5a\xae\xf7\x9b\x33\x6a\xfe\xec\xe9\x76\xfb\xb0\x6b\xbf\x10\x6b\xf9\xa7\xda\x17\x10\x92\x80\xe7\x8f\xb7\x25\x44\x9d\x04\x31\x9f\xd7\x62\x29\x7b\x64\x44\x24\x63\xf7\x85\x6d\x29\x71\x27\xa5\x55\xf7\x4d\xdb\x63\x48\x4c\x32\x9e\x3f\xde\x96\x90\xdc\x8a\xaa\x22\x43\x6f\xd7\x5f\xee\xb7\xdb\x26\xb7\xad\x5a\x34\x9f\xd4\xf6\xd3\xed\xa6\xdd\xd4\xdc\x35\xb8\x7a\xf3\x9b\xff\xfe\xb7\x0f\xfe\xdb\x37\xef\x3e\xfc\xe4\x79\xf7\x42\xfe\x57\xdc\x29\x7f\x51\xdf\xb5\x62\x5d\x37\xcb\x95\xe7\xd5\x0b\x28\xe1\x4f\xbd\xc9\x45\xf9\x65\xad\x56\x17\xde\xe4\x42\x36\x8b\xfb\x56\xad\x56\xd7\x77\x5f\xeb\x7b\xdc\xd0\x8b\x35\xfe\xea\x86\x7e\xaf\xeb\x66\xb3\xae\xe7\xb8\x68\x6c\x83\x7b\xb1\xfe\xf3\x5a\xd7\x73\x85\x13\xdc\x58\xad\xdb\x7a\x79\x67\x9f\xad\xeb\x85\xba\xf0\x66\x9e\xa7\x37\x4b\xd9\xa9\xf7\x2f\x25\xaa\x29\x4e\xfc\x7f\xff\x07\xdd\x5e\xfa\x4b\xb1\x50\x3e\x35\x9b\xf9\xd3\xee\xae\x6a\xdb\xa6\x9d\xf9\xdf\xbc\xc9\xdd\x57\x7b\xe5\xdf\xbc\xf2\xa1\xd5\xd5\x7b\xf5\x19\x42\x54\x3b\xb5\x6a\xe3\xfa\xd7\x8d\xd6\xaa\xb5\x62\x67\x33\x6f\x52\x6b\xdb\xe0\xa7\x57\xfe\xb2\x9e\x43\xc4\xa4\x55\xeb\x4d\xbb\xc4\xe5\xa5\xaf\x17\xeb\xab\xb7\x90\xae\xa7\x17\x10\xe4\xff\xfc\xf1\xc6\xff\xf9\xd3\x05\x69\x62\xfb\x9a\x79\x93\x07\xcf\x9b\x7c\x12\xad\x5f\x6e\xb4\x4f\xfd\x50\x27\xde\xe4\x96\xd4\x79\xe5\xd7\xcd\xd5\xeb\xe6\xfe\xcb\xf4\x97\x72\xa3\x2f\xfd\xbb\xaf\x33\x6f\x22\xe7\x6f\x3b\x4d\xaf\x5e\xcf\x9b\x95\x9a\xce\xbc\x73\xe9\x03\x31\x24\x7f\x40\x90\x6a\x5b\xd2\xdb\xdd\x2c\x37\xfa\xea\x57\xa8\x3e\x9d\x5d\xe2\x0d\xef\xc1\xf3\xe0\x35\xbe\x58\xad\xd4\x1a\x43\xbe\x91\x6b\x48\xb1\xf6\xb9\xf9\xf0\x26\xf5\x52\x37\xbe\xdf\xac\xae\xfe\x51\xcf\xd5\xbb\xa5\x6e\x1e\xdb\xb9\x29\xec\xee\x6f\x49\xb0\x73\xe8\xfb\x6e\x1a\xbd\xc9\xaa\xfe\x6a\xaf\xeb\xe5\x3a\x8d\xbd\xc9\x02\x50\xe2\x3f\x0a\xfd\x67\x53\x29\x7b\xf3\x43\xbd\x50\x3e\xdc\xe4\x0a\x67\xe8\xc7\xba\xca\x54\xd7\xbb\x7d\xcd\xfc\xf7\x62\xa1\xa6\x33\xd7\x03\xfa\x74\x56\xea\xfa\x0a\xbd\x7b\x0f\x23\x6d\x7f\xaf\xbf\xa2\xad\xd5\xe6\x79\x53\x28\x3a\xda\x14\xba\x4e\x67\xdb\x9a\x3f\x17\x00\xd3\x0e\x09\x80\x71\xd3\xd9\x93\xa1\x7b\x12\x9c\xf5\xc3\x42\xde\xad\xde\xd4\xed\x74\xe6\x97\x4d\x33\xdf\x6e\x2d\xe6\xab\x03\x96\x7f\x59\x91\xe1\xaa\xd5\x42\xaa\x6f\x0f\x5b\xad\x9d\x4b\xc0\xcb\x6f\x6f\xfb\x00\xfc\x8f\xfb\xdf\x3f\xce\xfd\x57\xce\x33\xa6\x17\xdc\x30\xcd\x4d\x5e\x72\x13\xe4\xdc\x04\x41\xff\xa1\x35\x37\x59\xc8\x4d\xa0\xe8\x3f\x4b\xb8\x09\x24\x37\x19\xe3\x26\x49\xa8\x2d\xce\xb3\x98\x1b\x1d\x3e\xdd\xd7\x01\xdd\x4b\x22\x3a\xd7\x11\x37\x81\xe6\x26\xc1\xbb\x01\x37\x79\x45\xf7\x21\x0b\xb2\x83\x82\x9b\x24\xe5\x26\xcc\xb9\x09\x2b\x6e\x62\x45\xff\xa1\xa0\xfb\x55\xc0\x8d\xc2\x01\x5d\x62\x6e\x22\x46\xfd\x31\xf4\x19\x51\x3f\x2c\xe1\x26\x0e\xb9\x49\x24\x37\x61\x44\xe7\x32\x20\x19\x5a\xd2\x75\x56\x71\x93\x49\x6e\x58\x4c\xd7\xba\xe2\x26\x82\x0e\x78\x1f\x72\x2a\xb2\xb1\xcc\xb9\x89\x70\x44\xd4\x07\xc6\x09\xe7\x49\x48\x7d\x05\xe1\xd3\x7d\x1c\xb2\x24\x1b\x71\x6e\xdf\x09\x68\x1c\x30\x1e\x69\x41\xef\xdb\x63\x6b\x5c\x71\x28\xc6\x4d\x01\xfb\x32\x6e\xaa\x90\x9b\x34\xdc\x1e\xff\x8b\x0e\x6f\x87\x67\xd4\x61\x42\x1f\xd6\x76\xc8\xb1\x85\xd5\xde\x64\x32\xe2\x1d\x97\xde\x64\x72\x31\x42\xfe\x17\x97\xde\x64\xf6\x18\xd8\xc3\x72\xa0\xce\xdf\x2c\x2e\x6d\xab\x63\x81\xe9\x11\xfd\x0f\x5a\x74\x08\x69\x1f\x01\xd2\x42\xdc\xcd\xab\xdd\x70\xf9\x06\x20\xb9\xf1\xc7\xed\xf1\x01\x19\x37\x7e\x91\x5f\xfa\x88\xfd\x9b\x6d\x68\x98\xc6\x51\x3a\xb3\xf7\x11\xd1\x37\x14\xf1\x7f\x2c\x6b\x33\x65\x09\x8b\x8b\x30\xc9\xd0\x2c\x98\x3d\x78\x13\x81\xfe\x7f\xb1\x16\x7f\xb3\x66\xde\xf8\xce\x5a\x28\x77\x63\x7f\x1f\x1e\x27\x44\x5c\xee\x46\xeb\x6e\x3e\xf5\xa6\xf9\xbc\xfc\xae\x68\x2d\x28\x3a\x6c\xc4\x05\xc3\x51\xda\x17\x6d\x65\xca\x4d\x90\x52\x94\xed\x7a\x6b\x59\x71\x93\x56\xdc\xc8\x84\x22\x98\xb1\x21\x6f\x1d\xb0\xe8\x24\x6f\x1d\x90\xe5\xbc\x75\x30\x17\xdd\xf3\xd6\x01\x39\x47\x7a\xeb\xb8\x45\xe7\xf3\xd6\x11\x7b\x9c\xb7\xb2\xec\x25\xbd\xb5\x37\xbd\x3f\x9d\x5c\x00\xbe\x78\x2f\xae\xb8\x09\x04\x37\x32\xe4\x26\x0a\xb8\x61\x20\x89\x84\x80\xb1\x28\x08\x74\x4b\xc9\x8d\x70\x20\x1f\x6a\x72\xbf\x14\xae\x18\x73\x93\x06\x04\xc8\x71\x49\x2e\x8e\x7b\x55\x4a\x80\x2e\xe1\xa6\x8c\x9b\x98\x71\x23\x24\xbd\x9b\x47\x8e\x50\xe0\xfe\x90\x19\x73\x53\x02\x84\x35\x37\x52\x53\x18\x88\x9c\x1b\x91\x71\x93\x23\x64\x14\x37\xcc\x91\x4c\xca\xb8\xc9\x33\x22\x92\xd0\x11\x1d\xec\x2c\x12\x7a\xae\x53\x6e\x4a\xc8\x2b\x28\xac\x92\x9c\x9b\xbc\xe0\xa6\x2c\xb9\xa9\x04\x37\x01\xfe\x63\x6e\x72\xc6\x8d\xca\x49\xe7\x38\xe7\xa6\x54\xdc\xe4\x92\x9b\x2a\xe7\x46\x95\xf4\x8f\x76\x45\x49\x7a\xe0\x1f\x04\xa6\x0b\x6e\x54\xc6\x8d\xc4\x78\x65\xdc\x30\x10\x55\x49\x61\x0d\x22\x52\xd0\x41\x71\x93\xe6\xf4\x6e\x9a\x71\x13\x09\x7a\x8e\x76\xa2\xe0\x86\x15\xd4\xae\x08\x89\x60\xb5\x20\x9d\x40\x62\x1a\xba\x69\x1a\x5b\x10\xa4\x25\xe8\x1d\x28\x00\xac\xc0\x16\xd8\x60\x49\xb5\xea\x87\x82\x61\x77\x39\x01\x0b\x86\x85\x59\x30\x18\xab\x3c\x77\xd0\x60\x58\xd2\x51\x70\x70\xd0\xaa\x73\xe1\xc1\xb8\x49\x1d\x20\x84\xc9\xcb\x22\xc2\x5e\x41\xff\xa3\x08\x2c\xee\x08\x0c\x29\x18\x08\xac\xe4\x26\x2c\x86\x09\x2c\xcc\x28\x3a\x10\x55\x69\x3c\x4c\x60\xc3\x26\x9d\xe4\xb5\x43\xc2\x9c\xd7\x0e\xaf\x87\xec\x79\xed\x90\xa4\x23\xbd\xf6\x80\x55\xe7\xf3\xda\x31\x93\xfe\x2f\x68\xac\x7f\x95\xe9\x74\x1e\x4b\xc1\x63\x8a\x78\x2c\x05\x2e\x47\x4f\x3c\x16\x6f\x25\xf8\xf0\xce\x34\xa5\xa2\x00\x1c\x14\x2a\x6e\x04\x23\xbe\x82\x57\xe3\xfd\x82\x11\xcf\x81\x5b\xac\xbc\x82\xda\x16\xce\xc3\x11\x15\x11\x22\x01\xfc\xa4\x89\x27\x10\x35\xe0\x39\x70\x18\xb8\x11\x72\x75\xc2\x8d\x48\x5d\xd4\x80\x4f\x70\x2f\xa3\xf7\xc0\x17\x41\xe6\x78\x08\xc5\x50\x40\x1c\xd0\x45\x45\x8c\xa2\x28\xe1\x46\x3a\xde\x03\x57\x56\x15\xf1\x64\x2e\xc8\x36\xe8\x10\x15\x8e\x93\x02\x2a\x70\xc0\xd9\x28\xae\x18\xf4\xd6\x24\x37\xcb\x29\xf2\x50\x54\x41\x26\xb8\x1c\xfc\x86\x36\x79\x40\x36\x45\x8a\x9b\x22\xa6\xb6\x28\xd6\x90\x62\xa6\x11\x71\x22\x03\x97\x29\x6e\x4a\xf0\xb2\xe0\x86\x65\xdc\xa4\xa5\xe3\xdb\x98\x9b\x42\x53\x71\x06\xdb\xc1\xdd\x85\xe3\x5e\x8c\xa1\xe5\xc9\x80\xf4\xc0\x58\xc8\xd4\x71\xad\x20\xd9\x15\x23\x5e\x05\x0f\x0a\xe4\x0f\xb0\xa7\xa0\x34\x17\xe3\x91\x21\x1f\xc0\xdc\x94\xd4\x3f\xd0\x07\xba\xc3\x76\xe4\x2a\xb2\xa0\x67\xb9\xe3\x76\xcc\xef\xe3\x1c\x2b\xb2\xa1\x10\x74\x0d\x54\x52\x11\xf9\x92\xcd\x5f\x72\xca\x2b\xe2\x8c\xfa\x49\x0a\x42\xa7\xc0\xb5\xb1\x32\xc1\xfd\x11\x8d\x39\x8a\x5a\x70\x37\xe6\x0b\xe3\x86\x5c\x23\x70\xf3\x0b\x8e\x47\xee\x60\xe7\x01\xed\x33\xb2\x07\xfd\x89\x8a\xc6\x04\xb2\xb3\x6c\xdf\x77\x71\x60\x2c\x31\x36\x18\xb3\xc8\xa1\x6c\x1f\x32\x8e\xc4\xcd\x09\xd0\x38\x22\xcd\x62\xe3\xe8\x5a\xf0\x0e\x38\x8e\xc8\x3a\x0a\x1d\x0f\x5b\x76\x2e\x78\x3c\x60\x95\xc3\xc7\x30\x0e\x5f\x16\x20\xf7\x57\xd9\x7f\x14\xad\xcb\x9c\x80\x2f\x94\x14\x90\x00\xa1\x8e\xea\x87\xa8\x5d\xb9\x67\x28\x02\xa2\x84\xfe\xfb\x1d\x78\xd0\xae\x93\x1c\x78\x50\x9a\x73\xe0\x91\xbd\x8a\x3d\x07\x1e\x94\x75\xa4\x03\x1f\xb2\xec\x7c\x0e\x3c\x6a\x55\x47\xf0\x2f\xba\xaa\x32\xb4\x09\xf4\x1d\x14\x1f\x13\xc4\xa2\x8c\x44\xa9\x9a\x77\x14\x1f\x13\x2d\x82\xe2\x55\x4a\xd7\x65\x44\x30\x0a\xda\x0c\x41\x79\x31\xc1\xab\xf5\x4a\x49\x91\xd0\xad\xa0\x28\x49\xf4\x8e\x72\x0e\x65\x65\x8c\x52\x2c\xa1\xb2\x14\xd0\x8d\x72\x11\x54\x9a\x80\xfe\x02\x2a\x2d\x6d\x1f\x82\x52\x02\xd0\x95\x72\x10\x6e\xcb\x31\x50\x50\x4c\x6b\x89\x89\xa2\xf2\x13\x54\x05\x1d\xb2\x80\xa8\x36\x77\xe9\x04\x52\x0b\xbc\x07\x5b\x40\x45\xa0\x1d\xe8\x69\xcb\x5d\x97\x24\x5b\x3d\x04\x51\x24\x68\x15\x54\x85\x92\x10\x74\x9b\x80\x22\x53\xb2\x1f\xff\x1a\xa9\x4b\x40\x6b\xb0\xa0\x68\xd0\x28\x68\xb7\x70\xd1\x1d\xba\x6b\x44\x77\xe2\x52\x03\xd8\x56\x74\xe5\xb1\x26\x2a\xab\x0a\x1a\x4b\x51\xd2\x12\x01\xca\x59\xd0\x1e\xc6\x08\x69\x0b\xf4\x2c\xdc\x9a\xad\x74\x05\x00\xe8\x1a\xf4\x88\xf4\x0b\x08\x81\xb2\xd6\xae\x55\xc7\x34\x07\xac\xe4\x86\x49\xa2\x53\x21\xb9\x51\x85\x5b\xef\x4d\x69\x2c\xd0\x56\xc4\xb4\xfe\x8b\xbe\x2b\xe9\x96\x12\x4a\xa2\x6f\xe8\x0a\xfa\x2c\x13\x9a\x2f\xa4\x5e\x7d\x54\x0a\x6a\x4e\x18\xad\x96\x21\x95\x29\x65\x3f\x12\x8d\x7a\xe8\x09\x58\x34\x2a\xcf\xa2\xd1\x81\x8d\xd1\x1d\x3c\x1a\x95\x77\x14\x22\x1d\x63\xe1\xb9\x30\xe9\xa0\x6d\x8f\xa8\xf4\xc2\xb0\xd4\xb7\xf1\xfc\x23\xeb\x65\x29\x29\x6b\xc5\x81\x2c\x1f\xd7\x76\x05\x4a\x8d\x13\x2c\xb2\x6e\xc5\x28\x93\xae\xdc\x16\x4a\xbf\x5b\x8f\x58\x78\x92\x5b\x8f\xc8\x73\x6e\x3d\xba\x99\xbf\xe7\xd6\x23\xf2\x8e\x74\xeb\xc3\x16\x9e\xcf\xad\x0f\xd8\xd6\x65\x8b\xc1\x4b\x7a\x75\xff\xb7\x12\xdf\x5f\x4d\xa3\x52\x05\x9c\xe7\xe9\xd6\xaa\x70\x4c\x55\x1e\x20\x1a\x95\x96\x06\x7c\xe7\xe4\x97\xa0\x3b\xbb\xcd\xa7\x69\xf5\x77\x88\x6e\x23\x46\x15\x11\xaa\xd5\x22\x25\xaa\x48\x5c\x15\x08\xba\xd4\x6e\x65\xd8\xae\x4c\x0b\x92\x0d\x3a\xb3\x95\x35\xa3\x98\xc9\x24\x55\xd1\x91\xdb\xda\xcc\x1d\x35\xa0\x9a\xb3\xdb\x9f\x09\x51\x0d\xfa\x04\xf5\x82\xc6\x2d\xb5\x0a\xaa\x48\x41\x11\xa8\xd0\xed\xaa\x6f\x46\xb1\x59\x39\x9a\x87\x7e\x76\xad\x2a\x20\xaa\x8a\x5c\xb5\x8a\x98\xc3\x98\xa0\xaa\x53\x8e\x8e\xa4\xa3\x39\x50\x31\xe2\xd8\x6e\x97\x32\xd2\x09\xd4\x68\x2b\xe1\x88\xfa\x02\x8d\x21\xe6\x61\x83\xad\xa6\x15\xad\xf0\x82\x72\xed\xea\xba\x4b\x61\xb2\x88\xde\x49\x73\xea\x1f\xe9\x08\xc6\x05\x95\x32\x52\x1b\x50\x6f\xe8\xf0\x23\xca\x88\xa2\x51\x5d\x16\xc0\x86\x88\xe8\x9d\x55\x94\x2e\xc1\x5e\x54\xf0\x78\x86\xaa\x14\x74\x1b\xb8\xea\x1c\x98\x83\xd4\x04\x73\x10\x86\x94\xb6\xa0\x7a\xb5\xfd\x56\x44\xef\x56\xef\x98\xd2\x0a\x19\x11\x0e\x61\x9e\x31\xef\x18\x53\xa4\x34\x55\xee\xaa\xf7\x88\x56\xb0\x31\xbe\xb6\x4f\x97\x9a\xd8\xed\x62\x46\x73\x0a\xaa\x87\x9c\xd8\x6d\x01\x97\xa8\x92\x05\xcd\x67\xe2\xae\x43\xb7\x56\x88\x94\x04\xbe\xa5\x73\x4a\x3b\xa0\x13\x7c\x0b\xb2\xa2\x90\xe6\x07\xa9\x14\xd2\x2f\xe8\xcf\xdc\x76\x32\x8b\x68\x7c\x03\xb6\x8f\xa3\xda\xed\x42\x60\x9c\xe1\x93\xf0\xab\xa7\xf7\x9e\x70\x74\x24\xa6\x4e\x40\xd1\x11\x69\x16\x43\x47\xbf\x76\xda\x41\xd0\x11\x59\x47\xe1\xe7\x61\xcb\xce\x85\x9e\x07\xac\x72\xd8\x19\xa5\xd1\xcb\x82\xe7\xfe\x77\x64\x3f\x72\x07\x18\x00\x07\xe2\x97\x6e\x59\x2a\x74\xcb\x58\x63\x3b\xc1\x68\x8b\x1c\x1f\x7d\x21\xb0\xe1\xf0\x7d\xc9\xc0\x88\x6d\x27\x39\xf1\xa0\x34\xe7\xc4\x23\x5f\xe4\xed\x39\xf1\xa0\xac\x23\x9d\xf8\x90\x65\xe7\x73\xe2\x51\xab\xba\xbc\xb6\x78\x49\x1f\xde\xfd\xd4\xf1\xfb\xbe\x36\xd2\x31\x79\x30\xbc\x34\xa8\x08\x96\xff\xea\x16\x10\x68\x30\x73\x9b\xb3\xf6\x5d\xf7\xe5\x11\x68\x43\x67\x2e\x42\x12\x3a\x50\xfd\xe6\x39\xa5\x16\xa0\x6a\x96\x52\x25\x88\x48\x40\xd5\x87\xca\x32\x77\x5f\x05\xd9\x85\xe3\x92\x22\xc3\xa6\x1f\x6e\x93\x15\x14\x1e\x3a\xaa\xc0\xbb\x68\xa3\x3b\x5b\x7b\xbe\xfc\x41\x9a\x00\x7d\x05\x23\x1a\x8e\xf2\xfe\x08\xea\x1d\xd9\x13\x62\xa7\x57\x8e\x8d\x9a\x81\xcf\x54\x77\xe2\xa5\xb7\xfd\x51\x91\x32\x66\xc1\xb9\x62\x64\xd0\x06\x17\x1d\x49\xfa\x57\xa3\x23\x61\x51\xc2\x92\xf8\x4c\xd1\xb1\xf7\xb5\xef\xf7\x02\xfc\x39\x43\x64\x97\x2c\xc6\xc2\x24\x70\xee\xbd\x1d\x26\x43\x1f\xb7\x21\x04\x58\xf7\xfd\x83\xa2\xf0\xe8\x77\xf1\xa1\xe1\x39\xc9\xcf\x87\x84\x39\x67\x1f\xfe\xf0\x7a\xcf\xe3\x87\x24\x1d\xe9\xf6\x07\xac\x3a\x9f\xef\x8f\x99\xe4\xaf\xea\xff\x05\x00\x00\xff\xff\x38\x52\xed\x4b\x00\x30\x00\x00")

func bindataGoBytes() ([]byte, error) {
	return bindataRead(
		_bindataGo,
		"bindata.go",
	)
}

func bindataGo() (*asset, error) {
	bytes, err := bindataGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata.go", size: 20480, mode: os.FileMode(436), modTime: time.Unix(1515135217, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"0_create_users_table.up.sql": _0_create_users_tableUpSql,
	"0_drop_users_table.down.sql": _0_drop_users_tableDownSql,
	"1_create_events_table.up.sql": _1_create_events_tableUpSql,
	"1_drop_events_table.down.sql": _1_drop_events_tableDownSql,
	"2_create_matches_table.up.sql": _2_create_matches_tableUpSql,
	"2_drop_matches_table.down.sql": _2_drop_matches_tableDownSql,
	"3_create_alliances_table.up.sql": _3_create_alliances_tableUpSql,
	"3_drop_alliances_table.down.sql": _3_drop_alliances_tableDownSql,
	"4_create_reports_table.up.sql": _4_create_reports_tableUpSql,
	"4_drop_reports_table.down.sql": _4_drop_reports_tableDownSql,
	"5_add_event_type.up.sql": _5_add_event_typeUpSql,
	"5_remove_event_type.down.sql": _5_remove_event_typeDownSql,
	"bindata.go": bindataGo,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"0_create_users_table.up.sql": &bintree{_0_create_users_tableUpSql, map[string]*bintree{}},
	"0_drop_users_table.down.sql": &bintree{_0_drop_users_tableDownSql, map[string]*bintree{}},
	"1_create_events_table.up.sql": &bintree{_1_create_events_tableUpSql, map[string]*bintree{}},
	"1_drop_events_table.down.sql": &bintree{_1_drop_events_tableDownSql, map[string]*bintree{}},
	"2_create_matches_table.up.sql": &bintree{_2_create_matches_tableUpSql, map[string]*bintree{}},
	"2_drop_matches_table.down.sql": &bintree{_2_drop_matches_tableDownSql, map[string]*bintree{}},
	"3_create_alliances_table.up.sql": &bintree{_3_create_alliances_tableUpSql, map[string]*bintree{}},
	"3_drop_alliances_table.down.sql": &bintree{_3_drop_alliances_tableDownSql, map[string]*bintree{}},
	"4_create_reports_table.up.sql": &bintree{_4_create_reports_tableUpSql, map[string]*bintree{}},
	"4_drop_reports_table.down.sql": &bintree{_4_drop_reports_tableDownSql, map[string]*bintree{}},
	"5_add_event_type.up.sql": &bintree{_5_add_event_typeUpSql, map[string]*bintree{}},
	"5_remove_event_type.down.sql": &bintree{_5_remove_event_typeDownSql, map[string]*bintree{}},
	"bindata.go": &bintree{bindataGo, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

