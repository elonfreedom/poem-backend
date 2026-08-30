-- 回滚：清空 poems.author_id
UPDATE poems SET author_id = NULL WHERE author_id IS NOT NULL;

-- 回滚：将 authors.dynasty 重置为 '未知'
UPDATE authors SET dynasty = '未知' WHERE dynasty != '未知';
