// HTTP服务 - 处理与后端的HTTP通信

const API_BASE_URL = 'http://localhost:8080'

/**
 * 搜索答案
 * @param {string} query - 搜索查询
 * @param {Object} filters - 过滤条件
 * @returns {Promise<Array>} 搜索结果
 */
export async function searchAnswers(query, filters = {}) {
  try {
    const response = await fetch(`${API_BASE_URL}/api/search`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query,
        filters
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP请求失败: ${response.status} ${response.statusText}`)
    }

    const data = await response.json()
    
    if (!data.success) {
      throw new Error(data.message || '搜索失败')
    }

    return data.results || []
  } catch (error) {
    console.error('搜索答案失败:', error)
    throw error
  }
}

/**
 * 测试HTTP连接
 * @returns {Promise<boolean>} 连接是否成功
 */
export async function testConnection() {
  try {
    const response = await fetch(`${API_BASE_URL}/api/search`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query: 'test',
        filters: {}
      })
    })

    return response.ok
  } catch (error) {
    console.error('连接测试失败:', error)
    return false
  }
}
