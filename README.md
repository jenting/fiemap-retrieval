### Fiemap retrieval on sparse file with multiple extents

The [fiemap](https://www.kernel.org/doc/Documentation/filesystems/fiemap.txt) is an efficient way for userspace to get file extent mappings. Instead of block-by-block mapping, fiemap returns a list of extents.

- Generate sparse file with multiple extents.
  ```shell
  make
  ./generator
  ```

- Install fiemap tool.
  ```shell
  sudo apt install gcc make -y
  wget https://github.com/joshimoo/fiemap/archive/refs/tags/v0.2.0.tar.gz
  tar zxvf v0.2.0.tar.gz
  cd fiemap-0.2.0
  make
  sudo make install
  ```

- Use `ls -lsah` to see the sparse file actual used space as well as max space.
  Example below 1GB allocated, 514MB used.
  ```shell
  514M -rw-r--r-- 1 jenting users 1.0G Mar 17 08:36 /tmp/  ssync-src-fiemap-1gb-file
  ```

- Then run `fiemap`.
  ```shell
  fiemap /tmp/ssync-src-fiemap-1gb-file
  ```
  If the initial 1024 retrieval takes around 3 minutes you can cancel it via ctrl+c this will take another 3 minutes till the process returns from the kernel. This is a slow EXT4 extent retrieval -> Bad Kernel.
  If the initial 1024 retrievals takes 0 seconds than you are good and can let it run to completion. -> Good Kernel.

  Example of a full run on a newer 5.8 kernel
  ```shell
  retrieved 1024 extents in 0 seconds

  fiemap done retrieved 131072 extents in 39 seconds
  File /tmp/ssync-src-fiemap-1gb-file has 131072 extents:
  ```
