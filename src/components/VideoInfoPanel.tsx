import { useState } from "react";
import type { TagItem, VideoDetail } from "@/types";
import { formatCount } from "@/lib/format";

type Props = {
  video: VideoDetail;
  availableTags?: TagItem[];
  tagSaving?: boolean;
  onTagsChange?: (tags: string[]) => Promise<void>;
};

export function VideoInfoPanel({
  video,
  availableTags = [],
  tagSaving = false,
  onTagsChange,
}: Props) {
  const [editingTags, setEditingTags] = useState(false);
  const [draftTags, setDraftTags] = useState<string[]>(video.tags ?? []);
  const [tagError, setTagError] = useState("");
  const [descCollapsed, setDescCollapsed] = useState(true);

  function openTagEditor() {
    setDraftTags(video.tags ?? []);
    setTagError("");
    setEditingTags(true);
  }

  async function saveTags() {
    if (!onTagsChange) return;
    setTagError("");
    try {
      await onTagsChange(draftTags);
      setEditingTags(false);
    } catch (e) {
      setTagError(e instanceof Error ? e.message : "保存标签失败");
    }
  }

  return (
    <section className="info-panel" aria-label="视频信息">
      <header className="info-panel__header">视频详细参数</header>
      <div className="info-panel__body">
        <div className="info-row">
          <span className="info-row__label">发布时间</span>
          <span className="info-row__value">{video.publishedAt}</span>
        </div>

        <div className="info-row">
          <span className="info-row__label">发布作者</span>
          <span className="info-row__value">{video.author || video.category || "影视合集"}</span>
        </div>

        <div className="info-row">
          <span className="info-row__label">播放次数</span>
          <span className="info-row__value">{formatCount(video.views)} 次观看</span>
        </div>

        <div className="info-row">
          <span className="info-row__label">画面品质</span>
          <span className="info-row__value">{video.quality || "HD 1080P"}</span>
        </div>

        {video.sourceLabel && (
          <div className="info-row">
            <span className="info-row__label">来源网盘</span>
            <span className="info-row__value">{video.sourceLabel}</span>
          </div>
        )}

        <div className="info-row">
          <span className="info-row__label">视频时长</span>
          <span className="info-row__value">{video.duration || "未知"}</span>
        </div>

        {/* 标签行 - 满宽 */}
        <div className="info-row is-tags-row">
          <span className="info-row__label">视频标签</span>
          <div className="info-row__value">
            <div className="detail-tags">
              {(video.tags ?? []).map((t) => (
                <span key={t} className="tag-chip">
                  {t}
                </span>
              ))}
              {onTagsChange && (
                <button className="detail-tags__edit" onClick={openTagEditor}>
                  修改标签
                </button>
              )}
            </div>
            {editingTags && (
              <div className="detail-tag-editor">
                <div className="detail-tag-editor__grid">
                  {availableTags.map((tag) => (
                    <label key={tag.id} className="detail-tag-editor__item">
                      <input
                        type="checkbox"
                        checked={draftTags.includes(tag.label)}
                        onChange={() => setDraftTags(toggleTag(draftTags, tag.label))}
                      />
                      <span>{tag.label}</span>
                      {typeof tag.count === "number" && <em>({tag.count})</em>}
                    </label>
                  ))}
                </div>
                {tagError && <div className="detail-tag-editor__error">{tagError}</div>}
                <div className="detail-tag-editor__actions">
                  <button onClick={() => setEditingTags(false)}>取消</button>
                  <button onClick={saveTags} disabled={tagSaving}>
                    {tagSaving ? "保存中..." : "保存修改"}
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* 描述行 - Collapsible */}
        {video.description && (
          <div 
            className="info-row" 
            style={{ 
              gridColumn: "1 / -1", 
              borderTop: "1px dashed rgba(255, 255, 255, 0.06)", 
              paddingTop: "var(--space-4)" 
            }}
          >
            <span className="info-row__label">视频简介</span>
            <div className="info-row__value" style={{ position: "relative" }}>
              <p className={`description ${descCollapsed ? "is-collapsed" : ""}`} style={{ margin: 0 }}>
                {video.description}
              </p>
              {video.description.length > 120 && (
                <button 
                  className="description-toggle" 
                  onClick={() => setDescCollapsed(!descCollapsed)}
                  style={{ border: 0, padding: 0, marginTop: "6px" }}
                >
                  {descCollapsed ? "展开全部介绍 ↓" : "收起介绍 ↑"}
                </button>
              )}
            </div>
          </div>
        )}
      </div>
    </section>
  );
}

function toggleTag(tags: string[], label: string): string[] {
  return tags.includes(label)
    ? tags.filter((tag) => tag !== label)
    : [...tags, label];
}
