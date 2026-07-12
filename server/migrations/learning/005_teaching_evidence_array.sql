ALTER TABLE learning_teaching_interventions
    ALTER COLUMN evidence SET DEFAULT '[]'::jsonb;

UPDATE learning_teaching_interventions
SET evidence = '[]'::jsonb
WHERE jsonb_typeof(evidence) <> 'array';

COMMENT ON COLUMN learning_teaching_interventions.evidence IS '文档与检索依据列表';
