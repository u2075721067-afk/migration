'use client';

import { useEffect } from 'react';
import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/themes/prism.css';

interface CodeHighlightProps {
  code: string;
  language?: string;
  className?: string;
}

export function CodeHighlight({ code, language = 'json', className = '' }: CodeHighlightProps) {
  useEffect(() => {
    Prism.highlightAll();
  }, [code]);

  return (
    <pre className={`language-${language} ${className}`}>
      <code className={`language-${language}`}>
        {code}
      </code>
    </pre>
  );
}
