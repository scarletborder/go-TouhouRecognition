import asyncio
import csv
import random
import logging
from typing import List, Dict, Optional, Set
from urllib.parse import urljoin

import nodriver as uc
from bs4 import BeautifulSoup

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class THBWikiMusicCrawler:
    """THBWiki 原曲爬虫 - 使用 nodriver"""
    
    BASE_URL = "https://thwiki.cc"
    START_URL = "https://thwiki.cc/%E5%88%86%E7%B1%BB:%E5%8E%9F%E6%9B%B2"
    
    def __init__(self):
        self.browser = None
        self.music_list = []  # 存储歌曲列表
        self.music_info = []  # 存储歌曲详细信息
        self.processed_urls = set()  # 已处理的歌曲URL
        self.total_pages = 0
        self.current_page = 0
        
    async def init_browser(self):
        """初始化浏览器"""
        self.browser = await uc.start(
            headless=False,  # 设置为 True 可启用无头模式
            window_size=(1920, 1080),
            lang="zh-CN",
            browser_executable_path=r"C:\Users\cirno\AppData\Local\Chromium\Application\chrome.exe",
            browser_args=[
                '--disable-blink-features=AutomationControlled',
                '--disable-dev-shm-usage',
                '--no-sandbox',
                '--disable-gpu',
            ]
        )
        logger.info("浏览器初始化完成")
        
    async def close_browser(self):
        """关闭浏览器"""
        if self.browser:
            self.browser.stop()
            logger.info("浏览器已关闭")
            
    async def get_page_content(self, url: str, max_retries: int = 3) -> str:
        """获取页面HTML内容（带重试机制与直接JS获取DOM）"""
        for attempt in range(1, max_retries + 1):
            try:
                logger.info(f"正在加载页面 (尝试 {attempt}/{max_retries}): {url}")
                tab = await self.browser.get(url)
                
                # 页面跳转后给一点缓冲时间，等待页面状态稳定
                await asyncio.sleep(1.0)
                
                # 尝试等待内容元素
                try:
                    await tab.wait_for('#mw-content-text', timeout=15)
                except Exception:
                    pass  # 如果超时或报错，继续尝试直接获取HTML
                
                # 关键修复：改用 evaluate 直接从 JS 运行时获取 outerHTML，避开 CDP Node ID 冲突
                html = await tab.evaluate('document.documentElement.outerHTML')
                
                if html and len(html) > 500:
                    return html
                else:
                    raise ValueError("获取到的 HTML 内容为空或不完整")
                    
            except Exception as e:
                logger.warning(f"加载页面失败 (第 {attempt} 次): {url}, 错误: {e}")
                if attempt < max_retries:
                    await asyncio.sleep(attempt * 1.5)  # 退避等待
                else:
                    logger.error(f"彻底无法加载页面: {url}")
                    return ""
        return ""
            
    def parse_music_list_page(self, html: str) -> tuple:
        """解析歌曲列表页面"""
        soup = BeautifulSoup(html, 'html.parser')
        songs = []
        next_url = None
        
        list_container = soup.find('div', class_='mw-category mw-category-columns')
        if not list_container:
            list_container = soup.find('div', class_='mw-category')
            
        if not list_container:
            logger.warning("未找到歌曲列表容器")
            return songs, next_url
            
        song_links = list_container.find_all('a', href=True)
        for link in song_links:
            song_name = link.get_text(strip=True)
            href = link.get('href', '')
            
            if song_name and href.startswith('/') and not href.startswith('/分类:'):
                full_url = urljoin(self.BASE_URL, href)
                songs.append({
                    'name': song_name,
                    'url': full_url
                })
                
        logger.info(f"当前页面找到 {len(songs)} 首歌曲")
        
        pages_div = soup.find('div', id='mw-pages')
        if pages_div:
            next_link = pages_div.find('a', string='下一页')
            if not next_link:
                for link in pages_div.find_all('a', href=True):
                    if 'pagefrom' in link.get('href', '') and link.get_text(strip=True) == '下一页':
                        next_link = link
                        break
                        
            if next_link and next_link.has_attr('href'):
                next_url = urljoin(self.BASE_URL, next_link['href'])
                
        return songs, next_url
        
    def parse_music_detail_page(self, html: str, url: str) -> Dict:
        """解析歌曲详情页面"""
        soup = BeautifulSoup(html, 'html.parser')
        music_name = self._extract_music_name(soup)
        translate_names = self._extract_translate_names(soup)
        works_and_assets = self._extract_works_and_assets(soup)
        
        return {
            'music_name': music_name,
            'url': url,
            'translate_names': translate_names,
            'works_and_assets': works_and_assets
        }
        
    def _extract_music_name(self, soup: BeautifulSoup) -> str:
        """提取曲名"""
        title_elem = soup.find('h1', id='firstHeading')
        if title_elem:
            full_title = title_elem.get_text(strip=True)
            if ':' in full_title:
                return full_title.split(':', 1)[1]
            return full_title
        return ""
        
    def _extract_translate_names(self, soup: BeautifulSoup) -> List[str]:
        """提取所有译名"""
        translate_names = []
        info_table = self._find_info_table(soup)
        if not info_table:
            return translate_names
            
        rows = info_table.find_all('tr')
        for row in rows:
            cells = row.find_all('td')
            if len(cells) >= 2:
                header = cells[0].get_text(strip=True)
                value = cells[1].get_text(strip=True)
                
                if header in ('曲名', '译名', '英文译名'):
                    if value:
                        translate_names.append(value)
                elif header == '其他译名':
                    value_html = str(cells[1])
                    cell_soup = BeautifulSoup(value_html, 'html.parser')
                    for br in cell_soup.find_all('br'):
                        br.replace_with('\n')
                    text = cell_soup.get_text(strip=False)
                    items = [item.strip() for item in text.split('\n') if item.strip()]
                    translate_names.extend(items)
                elif header == '常见误译':
                    value_html = str(cells[1])
                    cell_soup = BeautifulSoup(value_html, 'html.parser')
                    for br in cell_soup.find_all('br'):
                        br.replace_with('\n')
                    text = cell_soup.get_text(strip=False)
                    text = text.replace('~~', '').replace('<i>', '').replace('</i>', '')
                    items = [item.strip() for item in text.split('\n') if item.strip()]
                    translate_names.extend(items)
                elif header == '作曲':
                    break
                    
        return translate_names
        
    def _find_info_table(self, soup: BeautifulSoup) -> Optional[BeautifulSoup]:
        """找到基本信息表格"""
        info_header = soup.find('h2', id='基本信息')
        if not info_header:
            info_span = soup.find('span', id='.E5.9F.BA.E6.9C.AC.E4.BF.A1.E6.81.AF')
            if info_span:
                info_header = info_span.find_parent('h2')
                
        if not info_header:
            for h2 in soup.find_all('h2'):
                if '基本信息' in h2.get_text():
                    info_header = h2
                    break
                
        if info_header:
            next_table = info_header.find_next('table', class_='wikitable')
            if next_table:
                return next_table
                
        return None
        
    def _extract_works_and_assets(self, soup: BeautifulSoup) -> List[Dict]:
        """提取出处和对应的文件URL"""
        results = []
        works = self._extract_works(soup)
        asset_urls = self._extract_asset_urls(soup)
        asset_urls.sort(key=lambda x: 0 if 'mp3' in x.lower() else 1)
        
        if works and asset_urls:
            for i, work in enumerate(works):
                if i < len(asset_urls):
                    results.append({'work': work, 'url': asset_urls[i]})
                else:
                    results.append({'work': work, 'url': ''})
        elif works:
            for work in works:
                results.append({'work': work, 'url': ''})
        elif asset_urls:
            for url in asset_urls:
                results.append({'work': '未知出处', 'url': url})
                
        return results
        
    def _extract_works(self, soup: BeautifulSoup) -> List[str]:
        """提取出现作品列表"""
        works = []
        works_header = soup.find('h2', id='出现作品')
        if not works_header:
            works_span = soup.find('span', id='.E5.87.BA.E7.8E.B0.E4.BD.9C.E5.93.81')
            if works_span:
                works_header = works_span.find_parent('h2')
                
        if not works_header:
            for h2 in soup.find_all('h2'):
                if '出现作品' in h2.get_text():
                    works_header = h2
                    break
                
        if works_header:
            table = works_header.find_next('table')
            if table:
                rows = table.find_all('tr')
                for row in rows:
                    cells = row.find_all('td')
                    if len(cells) >= 2:
                        work_cell = cells[1]
                        work_link = work_cell.find('a')
                        if work_link:
                            work_name = work_link.get_text(strip=True)
                            if work_name:
                                works.append(work_name)
                                
        return works
        
    def _extract_asset_urls(self, soup: BeautifulSoup) -> List[str]:
        """提取所有音乐文件URL"""
        urls = []
        audio_tags = soup.find_all('audio')
        for audio in audio_tags:
            src = audio.get('src', '')
            if src.startswith('https://upload.thbwiki.cc/'):
                urls.append(src)
                
        file_links = soup.find_all('a', href=True)
        for link in file_links:
            href = link.get('href', '')
            if href.startswith('https://upload.thbwiki.cc/') and (href.endswith('.mp3') or href.endswith('.ogg')):
                if href not in urls:
                    urls.append(href)
                    
        urls.sort(key=lambda x: 0 if 'mp3' in x.lower() else 1)
        return urls
        
    async def crawl_music_list(self) -> List[Dict]:
        """爬取所有歌曲列表"""
        current_url = self.START_URL
        all_songs = []
        
        while current_url:
            self.current_page += 1
            logger.info(f"正在爬取第 {self.current_page} 页: {current_url}")
            
            html = await self.get_page_content(current_url)
            if not html:
                break
                
            songs, next_url = self.parse_music_list_page(html)
            all_songs.extend(songs)
            
            current_url = next_url
            await asyncio.sleep(random.uniform(0.5, 1.0))
            
        logger.info(f"总共找到 {len(all_songs)} 首歌曲")
        return all_songs
        
    async def crawl_music_details(self, songs: List[Dict]) -> List[Dict]:
        """爬取所有歌曲的详细信息"""
        all_details = []
        total = len(songs)
        
        for i, song in enumerate(songs):
            url = song['url']
            
            if url in self.processed_urls:
                continue
            self.processed_urls.add(url)
            
            try:
                logger.info(f"正在处理 [{i+1}/{total}]: {song['name']}")
                
                html = await self.get_page_content(url)
                if not html:
                    continue
                    
                detail = self.parse_music_detail_page(html, url)
                
                self.music_list.append({
                    'music_name': detail['music_name'],
                    'music_url': url,
                    'translate_names': '|'.join(detail['translate_names']) if detail['translate_names'] else ''
                })
                
                if detail['works_and_assets']:
                    for work_asset in detail['works_and_assets']:
                        self.music_info.append({
                            'music_name': detail['music_name'],
                            'original_works': work_asset['work'],
                            'asset_url': work_asset['url']
                        })
                else:
                    self.music_info.append({
                        'music_name': detail['music_name'],
                        'original_works': '',
                        'asset_url': ''
                    })
                
                all_details.append(detail)
                await asyncio.sleep(random.uniform(0.3, 0.8))
                
            except Exception as e:
                logger.error(f"处理歌曲失败 {song['name']}: {e}")
                
        return all_details
        
    async def run(self):
        """运行爬虫"""
        try:
            await self.init_browser()
            songs = await self.crawl_music_list()
            await self.crawl_music_details(songs)
            self.save_to_csv()
            logger.info("爬取完成!")
        except Exception as e:
            logger.error(f"爬取过程中发生错误: {e}")
        finally:
            await self.close_browser()
            
    def save_to_csv(self):
        """保存数据到CSV文件"""
        if self.music_list:
            with open('music_list.csv', 'w', newline='', encoding='utf-8-sig') as f:
                writer = csv.DictWriter(f, fieldnames=['music_name', 'music_url', 'translate_names'])
                writer.writeheader()
                writer.writerows(self.music_list)
            logger.info(f"已保存 {len(self.music_list)} 条歌曲列表到 music_list.csv")
            
        if self.music_info:
            with open('music_info.csv', 'w', newline='', encoding='utf-8-sig') as f:
                writer = csv.DictWriter(f, fieldnames=['music_name', 'original_works', 'asset_url'])
                writer.writeheader()
                writer.writerows(self.music_info)
            logger.info(f"已保存 {len(self.music_info)} 条歌曲信息到 music_info.csv")


async def main():
    crawler = THBWikiMusicCrawler()
    await crawler.run()


if __name__ == "__main__":
    uc.loop().run_until_complete(main())