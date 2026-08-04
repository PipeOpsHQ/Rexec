import asyncio, os, time
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[2] / "sdk" / "python"))
from rexec import RexecClient

async def main():
    url = os.environ["URL"]
    token = os.environ["TOKEN"]
    async with RexecClient(url, token) as client:
        print("[py] list...")
        before = await client.containers.list()
        print("[py] list count", len(before))
        print("[py] create...")
        c = await client.containers.create(image="ubuntu", name=f"py-e2e-{int(time.time())}")
        print("[py] created", c.id, c.status, c.image)
        after = await client.containers.list()
        print("[py] list count", len(after))
        got = await client.containers.get(c.id)
        print("[py] get", got.id, got.status)
        sid = c.id
        for _ in range(30):
            s = await client.sandboxes.get(sid)
            if s.status == "running":
                break
            if s.status == "error":
                raise RuntimeError("sandbox error")
            await asyncio.sleep(1)
        print("[py] exec...")
        r = await client.sandboxes.exec(sid, "echo rexec-e2e && uname -s")
        print("[py] exec exit", r.exit_code, "out", (r.stdout or r.output or "")[:80])
        if r.exit_code != 0:
            raise RuntimeError(f"exec failed: {r.exit_code}")
        await client.containers.delete(c.id)
        print("[py] OK")

asyncio.run(main())
