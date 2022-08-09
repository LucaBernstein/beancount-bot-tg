import os
from telethon import TelegramClient

class TestBot:
    def __init__(self) -> None:
        # Manage at https://my.telegram.org
        self.api_id = os.getenv('TG_API_ID')
        assert self.api_id is not None
        self.api_hash = os.getenv('TG_API_HASH')
        assert self.api_hash is not None
        self.client = TelegramClient('anon', self.api_id, self.api_hash)
