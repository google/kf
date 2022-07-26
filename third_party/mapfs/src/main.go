// Mounts another directory while mapping uid and gid to a different user.  Extends loopbackfs.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/mapfs/mapfs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	debug := flag.Bool("debug", false, "")
	uid := flag.Int64("uid", -1, "")
	gid := flag.Int64("gid", -1, "")
	fsName := flag.String("fsname", "mapfs", "")
	autoCache := flag.Bool("auto_cache", false, "")

	flag.Parse()
	if flag.NArg() < 2 || *uid <= 0 || *gid <= 0 {
		fmt.Printf("usage: %s -uid UID -gid GID [-fsname FSNAME] [-auto_cache] [-debug] MOUNTPOINT ORIGINAL\n", path.Base(os.Args[0]))
		fmt.Printf("UID and GID must be > 0")
		os.Exit(2)
	}

	orig := flag.Arg(1)
	loopbackfs := pathfs.NewLoopbackFileSystem(orig)
	finalFs := mapfs.NewMapFileSystem(*uid, *gid, loopbackfs, orig, &syscallshim.SyscallShim{})

	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}

	fuseOpts := []string{}
	if *autoCache {
		fmt.Println("warning -- auto_cache flag ignored as it is unsupported in fusermount")
	}

	pathFs := pathfs.NewPathNodeFs(finalFs, &pathfs.PathNodeFsOptions{})
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	mountPoint := flag.Arg(0)
	if err := os.MkdirAll(mountPoint, os.ModePerm); err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}
	// Caveat: this is the vcap user. Need to revisit this if we are to support other users.
	os.Chown(mountPoint, 2000, 2000)

	origAbs, _ := filepath.Abs(orig)
	mOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       *fsName,
		FsName:     origAbs,
		Debug:      *debug,
	}
	if len(fuseOpts) > 0 {
		mOpts.Options = fuseOpts
	}
	state, err := fuse.NewServer(conn.RawFS(), mountPoint, mOpts)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mounted!")
	state.Serve()
}
