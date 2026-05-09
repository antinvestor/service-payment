--- Copyright 2023-2026 Ant Investor Ltd
---
--- Licensed under the Apache License, Version 2.0 (the "License");
--- you may not use this file except in compliance with the License.
--- You may obtain a copy of the License at
---
---      http://www.apache.org/licenses/LICENSE-2.0
---
--- Unless required by applicable law or agreed to in writing, software
--- distributed under the License is distributed on an "AS IS" BASIS,
--- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
--- See the License for the specific language governing permissions and
--- limitations under the License.

CREATE OR REPLACE FUNCTION jsonb_to_tsv(jdoc jsonb)
    RETURNS tsvector AS $$
DECLARE
    search_text text;
BEGIN
    IF jdoc IS NULL OR jsonb_typeof(jdoc) NOT IN ('object', 'array') THEN
        RETURN ''::tsvector;
    END IF;

    SELECT string_agg(value, ' ')
    INTO search_text
    FROM (
        WITH RECURSIVE all_values AS (
            SELECT value
            FROM jsonb_each_text(jdoc)

            UNION ALL

            SELECT jsonb_array_elements_text(ja.elem)
            FROM (
                SELECT value as elem
                FROM jsonb_each(jdoc)
                WHERE jsonb_typeof(value) = 'array'
            ) ja

            UNION ALL

            SELECT jt.value
            FROM (
                SELECT value as obj
                FROM jsonb_each(jdoc)
                WHERE jsonb_typeof(value) = 'object'
            ) jo,
            LATERAL (
                SELECT value FROM jsonb_each_text(jo.obj)
            ) jt
        )
        SELECT value FROM all_values
    ) AS all_text_values;

    RETURN to_tsvector('english', COALESCE(search_text, ''));
END;
$$ LANGUAGE plpgsql IMMUTABLE;