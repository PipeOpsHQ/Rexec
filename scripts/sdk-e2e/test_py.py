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
        await client.containers.delete(c.id)
        print("[py] OK")

asyncio.run(main())
