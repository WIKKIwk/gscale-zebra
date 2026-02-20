package app

import (
	"os"
	"reflect"
	"syscall"
	"testing"
)

func TestParseCalibrateOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		text    string
		want    calibrateOptions
		wantErr bool
	}{
		{
			name: "default",
			text: "/calibrate",
			want: calibrateOptions{save: true},
		},
		{
			name: "device no-save dry-run",
			text: "/calibrate --device /dev/usb/lp1 --no-save --dry-run",
			want: calibrateOptions{
				device: "/dev/usb/lp1",
				save:   false,
				dryRun: true,
			},
		},
		{
			name: "device equals syntax",
			text: "/calibrate --device=/dev/usb/lp2",
			want: calibrateOptions{
				device: "/dev/usb/lp2",
				save:   true,
			},
		},
		{
			name:    "unknown arg",
			text:    "/calibrate --oops",
			wantErr: true,
		},
		{
			name:    "missing device value",
			text:    "/calibrate --device",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCalibrateOptions(tc.text)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCalibrateOptions error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("options mismatch\ngot : %#v\nwant: %#v", got, tc.want)
			}
		})
	}
}

func TestBuildCalibrationCommands(t *testing.T) {
	t.Parallel()

	withSave := buildCalibrationCommands(true)
	wantWithSave := []string{"~JC\n", "^XA^JUS^XZ\n"}
	if !reflect.DeepEqual(withSave, wantWithSave) {
		t.Fatalf("with-save mismatch\ngot : %#v\nwant: %#v", withSave, wantWithSave)
	}

	noSave := buildCalibrationCommands(false)
	wantNoSave := []string{"~JC\n"}
	if !reflect.DeepEqual(noSave, wantNoSave) {
		t.Fatalf("no-save mismatch\ngot : %#v\nwant: %#v", noSave, wantNoSave)
	}
}

func TestFriendlyCalibrateError(t *testing.T) {
	t.Parallel()

	msg := friendlyCalibrateError(&os.PathError{Op: "open", Path: "/dev/usb/lp0", Err: syscall.ENOENT}, "/dev/usb/lp0")
	if msg != "Zebra qurilmasi ulanmagan yoki topilmadi: /dev/usb/lp0" {
		t.Fatalf("unexpected not-exist message: %q", msg)
	}

	msg = friendlyCalibrateError(syscall.EBUSY, "/dev/usb/lp0")
	if msg != "Zebra qurilma band, boshqa dastur ishlatyapti" {
		t.Fatalf("unexpected busy message: %q", msg)
	}
}
