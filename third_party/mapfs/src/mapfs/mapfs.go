package mapfs

import (
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"code.cloudfoundry.org/goshims/syscallshim"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

const (
	CURRENT_ID = -1
)

//go:generate counterfeiter -o ../mapfs_fakes/fake_file_system.go  ../vendor/github.com/hanwen/go-fuse/v2/fuse/pathfs FileSystem

type mapFileSystem struct {
	pathfs.FileSystem
	uid, gid      int64
	syscall       syscallshim.Syscall
	root          string
	disableXAttrs bool
}

func NewMapFileSystem(uid, gid int64, fs pathfs.FileSystem, root string, sys syscallshim.Syscall) pathfs.FileSystem {
	// Make sure the Root path is absolute to avoid problems when the
	// application changes working directory.
	root, err := filepath.Abs(root)
	if err != nil {
		panic(err)
	}

	return &mapFileSystem{
		FileSystem:    fs,
		uid:           uid,
		gid:           gid,
		syscall:       sys,
		root:          root,
		disableXAttrs: false,
	}
}

func (fs *mapFileSystem) setEffectiveIDs(euid, egid int) (ouid, ogid int, err error) {
	ouid = fs.syscall.Geteuid()
	ogid = fs.syscall.Getegid()
	if egid != ogid {
		if err := fs.syscall.Setregid(CURRENT_ID, int(fs.gid)); err != nil {
			return ouid, ogid, err
		}
	}
	if euid != ouid {
		if err := fs.syscall.Setreuid(CURRENT_ID, int(fs.uid)); err != nil {
			return ouid, ogid, err
		}
	}

	return ouid, ogid, nil
}

func (fs *mapFileSystem) getPath(relPath string) string {
	return filepath.Join(fs.root, relPath)
}

func (fs *mapFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
}

func (fs *mapFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	a, code = fs.FileSystem.GetAttr(name, context)

	if a != nil {
		if int64(a.Uid) == fs.uid {
			a.Uid = context.Uid
		}
		if int64(a.Gid) == fs.gid {
			a.Gid = context.Gid
		}
	}

	return a, code
}

func (fs *mapFileSystem) Chmod(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}
	return fs.FileSystem.Chmod(name, mode, context)
}

func (fs *mapFileSystem) Chown(name string, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	if uid == context.Uid {
		uid = uint32(fs.uid)
	}
	if gid == context.Gid {
		gid = uint32(fs.gid)
	}
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}
	return fs.FileSystem.Chown(name, uid, gid, context)
}

func (fs *mapFileSystem) Utimens(name string, Atime *time.Time, Mtime *time.Time, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Utimens(name, Atime, Mtime, context)
}

func (fs *mapFileSystem) Truncate(name string, size uint64, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Truncate(name, size, context)
}

func (fs *mapFileSystem) Access(name string, mode uint32, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.ToStatus(fs.syscall.Faccessat(0, fs.getPath(name), mode, unix.AT_EACCESS))
}

func (fs *mapFileSystem) Link(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Link(oldName, newName, context)
}

func (fs *mapFileSystem) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Mkdir(name, mode, context)
}

func (fs *mapFileSystem) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) fuse.Status {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Mknod(name, mode, dev, context)
}

func (fs *mapFileSystem) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Rename(oldName, newName, context)
}

func (fs *mapFileSystem) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Rmdir(name, context)
}

func (fs *mapFileSystem) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Unlink(name, context)
}

func (fs *mapFileSystem) GetXAttr(name string, attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	if fs.disableXAttrs == true {
		return nil, fuse.Status(syscall.ENOTSUP)
	}

	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	xAttrBytes, code := fs.FileSystem.GetXAttr(name, attribute, context)
	if code == fuse.Status(syscall.ENOTSUP) {
		fs.disableXAttrs = true
	}
	return xAttrBytes, code
}

func (fs *mapFileSystem) ListXAttr(name string, context *fuse.Context) (attributes []string, code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fs.FileSystem.ListXAttr(name, context)
}

func (fs *mapFileSystem) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.RemoveXAttr(name, attr, context)
}

func (fs *mapFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.SetXAttr(name, attr, data, flags, context)
}

func (fs *mapFileSystem) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fs.FileSystem.Open(name, flags, context)
}

func (fs *mapFileSystem) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fs.FileSystem.Create(name, flags, mode, context)
}

func (fs *mapFileSystem) OpenDir(name string, context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fs.FileSystem.OpenDir(name, context)
}

func (fs *mapFileSystem) Symlink(value string, linkName string, context *fuse.Context) (code fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fs.FileSystem.Symlink(value, linkName, context)
}

func (fs *mapFileSystem) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return "", fuse.ToStatus(err)
	}

	return fs.FileSystem.Readlink(name, context)
}

func (fs *mapFileSystem) StatFs(name string) *fuse.StatfsOut {
	_, _, err := fs.setEffectiveIDs(int(fs.uid), int(fs.gid))
	if err != nil {
		return nil
	}

	stats := fs.FileSystem.StatFs(name)
	if stats != nil {
		stats.Bfree = stats.Blocks
	}
	return stats
}
