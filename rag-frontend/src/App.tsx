import { useState } from "react";
import { Search, Sparkles, BookOpen, Copy, Check } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import ReactMarkdown from "react-markdown";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { atomDark } from "react-syntax-highlighter/dist/esm/styles/prism";

function App() {
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<any>(null);

  const handleSearch = async () => {
    if (!query) return;
    setLoading(true);
    setData(null);
    try {
      const res = await fetch(`http://localhost:8000/search?q=${encodeURIComponent(query)}`);
      const json = await res.json();
      setData(json);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-dark text-gray-200 p-8 flex flex-col items-center">
      <motion.div initial={{ y: -20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} className="w-full max-w-3xl">

        {/* HEADER */}
        <h1 className="text-4xl font-bold text-center mb-8 flex items-center justify-center gap-3">
          <Sparkles className="text-accent" />
          <span className="bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-emerald-400">
            GenUI Code Search
          </span>
        </h1>

        {/* SEARCH BAR */}
        <div className="relative group z-10">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
            placeholder="Ask your codebase..."
            className="w-full bg-card border border-border rounded-xl p-4 pl-12 text-lg focus:outline-none focus:ring-2 focus:ring-accent transition-all shadow-lg"
          />
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-500 group-focus-within:text-accent transition-colors" />
        </div>

        {/* LOADING STATE */}
        {loading && (
          <div className="mt-8 text-center animate-pulse text-accent">
            🧠 Reasoning with Gemeni...
          </div>
        )}

        {/* RESULTS AREA */}
        <AnimatePresence>
          {data && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              className="mt-8 space-y-6"
            >

              {/* 🤖 AI ANSWER CARD */}
              <div className="bg-card border border-border rounded-xl p-6 shadow-2xl overflow-hidden relative">
                <div className="absolute top-0 left-0 w-1 h-full bg-accent" />
                <h3 className="text-accent font-semibold mb-4 flex items-center gap-2">
                  <Sparkles size={18} /> AI Analysis
                </h3>

                <div className="prose prose-invert max-w-none">
                  <ReactMarkdown
                    components={{
                      code({ node, inline, className, children, ...props }: any) {
                        const match = /language-(\w+)/.exec(className || "");
                        return !inline && match ? (
                          <div className="relative group rounded-lg overflow-hidden my-4 border border-border">
                            <div className="absolute right-2 top-2 opacity-0 group-hover:opacity-100 transition-opacity">
                              <CopyButton text={String(children).replace(/\n$/, "")} />
                            </div>
                            <SyntaxHighlighter
                              style={atomDark}
                              language={match[1]}
                              PreTag="div"
                              {...props}
                            >
                              {String(children).replace(/\n$/, "")}
                            </SyntaxHighlighter>
                          </div>
                        ) : (
                          <code className="bg-gray-800 rounded px-1 py-0.5 text-blue-300" {...props}>
                            {children}
                          </code>
                        );
                      },
                    }}
                  >
                    {data.answer}
                  </ReactMarkdown>
                </div>
              </div>

              {/* 📚 SOURCES */}
              <div className="grid gap-4">
                <h4 className="text-gray-400 font-medium flex items-center gap-2 mt-4">
                  <BookOpen size={16} /> Reference Sources
                </h4>
                {data.results?.map((res: any, i: number) => (
                  <motion.div
                    key={i}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.1 }}
                    className="bg-dark border border-border rounded-lg p-4 hover:border-accent/50 transition-colors"
                  >
                    <div className="flex justify-between text-xs text-gray-500 mb-2 uppercase tracking-wider font-semibold">
                      <span>{res.Metadata?.filename || "Unknown"}</span>
                      <span className="text-accent">{res.Metadata?.type || "Text"}</span>
                    </div>
                    <pre className="text-sm text-gray-300 overflow-x-auto whitespace-pre-wrap font-mono bg-[#0d1117] p-2 rounded">
                      {res.Text}
                    </pre>
                  </motion.div>
                ))}
              </div>

            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  );
}

// Helper Component for Copy functionality
const CopyButton = ({ text }: { text: string }) => {
  const [copied, setCopied] = useState(false);

  const onCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={onCopy}
      className="p-1.5 bg-gray-700 hover:bg-gray-600 rounded-md text-white transition-colors"
      title="Copy Code"
    >
      {copied ? <Check size={14} className="text-green-400" /> : <Copy size={14} />}
    </button>
  );
};

export default App;
