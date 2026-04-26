import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch({ headless: false, slowMo: 300 });
  const page = await browser.newPage();

  page.on('response', async (response) => {
    const url = response.url();
    if (url.includes('/auth/')) {
      let body = '';
      try { body = await response.text(); } catch {}
      console.log(`[${response.status()}] ${url}`);
      console.log('  响应:', body.substring(0, 400));
    }
  });

  console.log('导航到注册页...');
  await page.goto('http://localhost:3000/register');

  // Wait for the specific input to appear (more reliable than networkidle)
  await page.waitForSelector('input[placeholder="请输入手机号"]', { timeout: 15000 });

  // 用新账号注册
  // 随机账号，避免重复注册冲突
  const suffix = Math.floor(Math.random() * 900000) + 100000  // 6位随机数
  const phone = `138${String(suffix).padStart(8, '0')}`
  const userName = `testuser_${suffix}`

  await page.getByPlaceholder('请输入手机号').fill(phone);
  await page.getByPlaceholder('字母/数字/_-，不能数字开头').fill(userName);
  console.log(`注册账号: ${userName} / ${phone}`);
  await page.getByPlaceholder('至少 8 位密码').fill('Modelcraft123!');
  await page.getByPlaceholder('请再次输入密码').fill('Modelcraft123!');

  console.log('提交注册...');
  await page.getByRole('button', { name: '注册' }).click();

  // 等待跳转
  try {
    await page.waitForURL('**/workspace**', { timeout: 12000 });
    console.log('注册成功，已跳转到:', page.url());
  } catch {
    const url = page.url();
    const error = await page.locator('.text-destructive').first().textContent().catch(() => null);
    console.log('当前 URL:', url);
    if (error?.trim()) console.log('错误信息:', error.trim());

    // 截图帮助调试
    await page.screenshot({ path: '/tmp/register-result.png' });
    console.log('截图已保存到 /tmp/register-result.png');
  }

  await browser.close();
})();
