import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Package,
  ChevronLeft,
  Edit2,
  Trash2,
  Save,
  X,
  DollarSign,
  Tag,
  MapPin,
  ShoppingBag,
  ExternalLink,
  RefreshCw,
  AlertCircle,
  Check,
  Copy,
} from 'lucide-react';
import { api } from '../api/client';
import { cn } from '../lib/utils';
import type {
  InventoryItem,
  ListingStatus,
  ItemCondition,
} from '../api/types';

export function InventoryDetail() {
  const { sku } = useParams<{ sku: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [isEditing, setIsEditing] = useState(false);
  const [editData, setEditData] = useState<Partial<InventoryItem>>({});
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [selectedImage, setSelectedImage] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const {
    data: item,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['inventory-item', sku],
    queryFn: () => api.getInventoryItem(sku!),
    enabled: !!sku,
  });

  const { data: images } = useQuery({
    queryKey: ['inventory-images', sku],
    queryFn: () => api.getInventoryImages(sku!),
    enabled: !!sku,
  });

  const updateMutation = useMutation({
    mutationFn: (updates: Partial<InventoryItem>) =>
      api.updateInventoryItem(sku!, updates),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory-item', sku] });
      queryClient.invalidateQueries({ queryKey: ['inventory-items'] });
      queryClient.invalidateQueries({ queryKey: ['inventory-summary'] });
      setIsEditing(false);
      setEditData({});
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => api.deleteInventoryItem(sku!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['inventory-items'] });
      queryClient.invalidateQueries({ queryKey: ['inventory-summary'] });
      navigate('/inventory/list');
    },
  });

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(value);

  const formatDate = (dateStr: string) =>
    new Date(dateStr).toLocaleDateString('en-US', {
      weekday: 'short',
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });

  const handleEdit = () => {
    setEditData({
      name: item?.name,
      description: item?.description,
      category: item?.category,
      subcategory: item?.subcategory,
      brand: item?.brand,
      model: item?.model,
      condition: item?.condition,
      listing_status: item?.listing_status,
      location: item?.location,
      notes: item?.notes,
      current_value: item?.current_value,
    });
    setIsEditing(true);
  };

  const handleSave = () => {
    updateMutation.mutate(editData);
  };

  const handleCancel = () => {
    setIsEditing(false);
    setEditData({});
  };

  const handleCopySKU = () => {
    navigator.clipboard.writeText(sku || '');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <RefreshCw className="w-6 h-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !item) {
    return (
      <div className="space-y-4">
        <Link
          to="/inventory/list"
          className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
        >
          <ChevronLeft className="w-4 h-4" />
          Back to List
        </Link>
        <div className="flex items-center gap-2 p-4 rounded-lg border border-red-500/30 bg-red-500/5 text-red-400">
          <AlertCircle className="w-5 h-5" />
          <span>
            {error instanceof Error ? error.message : 'Item not found'}
          </span>
        </div>
      </div>
    );
  }

  const allImages = images?.images || item.images || [];
  const displayImage = selectedImage || item.primary_image || allImages[0];

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-4">
          <Link
            to="/inventory/list"
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
          >
            <ChevronLeft className="w-4 h-4" />
            Back
          </Link>
          <div>
            <h1 className="text-xl font-bold">{item.name}</h1>
            <div className="flex items-center gap-2 mt-1">
              <button
                onClick={handleCopySKU}
                className="flex items-center gap-1 text-xs font-mono text-muted-foreground hover:text-foreground"
              >
                {item.sku}
                {copied ? (
                  <Check className="w-3 h-3 text-green-400" />
                ) : (
                  <Copy className="w-3 h-3" />
                )}
              </button>
              <StatusBadge status={item.listing_status} />
              <ConditionBadge condition={item.condition} />
            </div>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {isEditing ? (
            <>
              <button
                onClick={handleCancel}
                className="flex items-center gap-1 px-3 py-1.5 rounded-md border border-border hover:bg-accent text-sm"
              >
                <X className="w-4 h-4" />
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={updateMutation.isPending}
                className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 text-sm"
              >
                <Save className="w-4 h-4" />
                Save
              </button>
            </>
          ) : (
            <>
              <button
                onClick={handleEdit}
                className="flex items-center gap-1 px-3 py-1.5 rounded-md border border-border hover:bg-accent text-sm"
              >
                <Edit2 className="w-4 h-4" />
                Edit
              </button>
              {deleteConfirm ? (
                <>
                  <button
                    onClick={() => deleteMutation.mutate()}
                    disabled={deleteMutation.isPending}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-md bg-red-500 text-white hover:bg-red-600 text-sm"
                  >
                    <Check className="w-4 h-4" />
                    Confirm Delete
                  </button>
                  <button
                    onClick={() => setDeleteConfirm(false)}
                    className="flex items-center gap-1 px-3 py-1.5 rounded-md border border-border hover:bg-accent text-sm"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </>
              ) : (
                <button
                  onClick={() => setDeleteConfirm(true)}
                  className="flex items-center gap-1 px-3 py-1.5 rounded-md border border-red-500/30 text-red-400 hover:bg-red-500/10 text-sm"
                >
                  <Trash2 className="w-4 h-4" />
                  Delete
                </button>
              )}
            </>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
        {/* Image Gallery */}
        <div className="lg:col-span-1 space-y-3">
          <div className="aspect-square rounded-lg border border-border bg-card overflow-hidden">
            {displayImage ? (
              <img
                src={displayImage}
                alt={item.name}
                className="w-full h-full object-contain"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center bg-secondary">
                <Package className="w-16 h-16 text-muted-foreground" />
              </div>
            )}
          </div>
          {allImages.length > 1 && (
            <div className="grid grid-cols-4 gap-2">
              {allImages.map((img, i) => (
                <button
                  key={i}
                  onClick={() => setSelectedImage(img)}
                  className={cn(
                    'aspect-square rounded border overflow-hidden',
                    selectedImage === img || (!selectedImage && i === 0)
                      ? 'border-primary'
                      : 'border-border hover:border-primary/50'
                  )}
                >
                  <img
                    src={img}
                    alt={`${item.name} ${i + 1}`}
                    className="w-full h-full object-cover"
                  />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Details */}
        <div className="lg:col-span-2 space-y-4">
          {/* Purchase Info */}
          <div className="rounded-lg border border-border bg-card p-4">
            <h2 className="text-sm font-semibold mb-3 flex items-center gap-2">
              <DollarSign className="w-4 h-4" />
              Purchase Information
            </h2>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div>
                <div className="text-xs text-muted-foreground">
                  Purchase Price
                </div>
                <div className="text-lg font-semibold text-green-400">
                  {formatCurrency(item.purchase_price)}
                </div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">
                  Current Value
                </div>
                {isEditing ? (
                  <input
                    type="number"
                    step="0.01"
                    value={editData.current_value ?? item.current_value ?? ''}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        current_value: parseFloat(e.target.value) || undefined,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-lg font-semibold">
                    {item.current_value
                      ? formatCurrency(item.current_value)
                      : '—'}
                  </div>
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">
                  Purchase Date
                </div>
                <div className="text-sm">{formatDate(item.purchase_date)}</div>
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Source</div>
                <div className="text-sm capitalize">{item.purchase_source}</div>
              </div>
            </div>
            {item.order_id && (
              <div className="mt-3 pt-3 border-t border-border">
                <div className="text-xs text-muted-foreground">Order ID</div>
                <div className="text-sm font-mono">{item.order_id}</div>
              </div>
            )}
            {item.product_url && (
              <div className="mt-2">
                <a
                  href={item.product_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                >
                  <ExternalLink className="w-3 h-3" />
                  View Product Page
                </a>
              </div>
            )}
          </div>

          {/* Product Info */}
          <div className="rounded-lg border border-border bg-card p-4">
            <h2 className="text-sm font-semibold mb-3 flex items-center gap-2">
              <Tag className="w-4 h-4" />
              Product Information
            </h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="text-xs text-muted-foreground">Category</div>
                {isEditing ? (
                  <input
                    type="text"
                    value={editData.category ?? item.category}
                    onChange={(e) =>
                      setEditData({ ...editData, category: e.target.value })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-sm">{item.category}</div>
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Subcategory</div>
                {isEditing ? (
                  <input
                    type="text"
                    value={editData.subcategory ?? item.subcategory ?? ''}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        subcategory: e.target.value || undefined,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-sm">{item.subcategory || '—'}</div>
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Brand</div>
                {isEditing ? (
                  <input
                    type="text"
                    value={editData.brand ?? item.brand ?? ''}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        brand: e.target.value || undefined,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-sm">{item.brand || '—'}</div>
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Model</div>
                {isEditing ? (
                  <input
                    type="text"
                    value={editData.model ?? item.model ?? ''}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        model: e.target.value || undefined,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-sm">{item.model || '—'}</div>
                )}
              </div>
            </div>
            {(item.description || isEditing) && (
              <div className="mt-3 pt-3 border-t border-border">
                <div className="text-xs text-muted-foreground mb-1">
                  Description
                </div>
                {isEditing ? (
                  <textarea
                    value={editData.description ?? item.description ?? ''}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        description: e.target.value || undefined,
                      })
                    }
                    rows={3}
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm resize-none"
                  />
                ) : (
                  <div className="text-sm text-muted-foreground">
                    {item.description}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Status Info */}
          <div className="rounded-lg border border-border bg-card p-4">
            <h2 className="text-sm font-semibold mb-3 flex items-center gap-2">
              <ShoppingBag className="w-4 h-4" />
              Status & Location
            </h2>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div>
                <div className="text-xs text-muted-foreground">Condition</div>
                {isEditing ? (
                  <select
                    value={editData.condition ?? item.condition}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        condition: e.target.value as ItemCondition,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  >
                    <option value="new">New</option>
                    <option value="like_new">Like New</option>
                    <option value="good">Good</option>
                    <option value="fair">Fair</option>
                    <option value="poor">Poor</option>
                    <option value="for_parts">For Parts</option>
                  </select>
                ) : (
                  <ConditionBadge condition={item.condition} />
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">
                  Listing Status
                </div>
                {isEditing ? (
                  <select
                    value={editData.listing_status ?? item.listing_status}
                    onChange={(e) =>
                      setEditData({
                        ...editData,
                        listing_status: e.target.value as ListingStatus,
                      })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  >
                    <option value="not_listed">Not Listed</option>
                    <option value="pending_review">Pending Review</option>
                    <option value="listed">Listed</option>
                    <option value="sold">Sold</option>
                    <option value="keeping">Keeping</option>
                  </select>
                ) : (
                  <StatusBadge status={item.listing_status} />
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Location</div>
                {isEditing ? (
                  <input
                    type="text"
                    value={editData.location ?? item.location}
                    onChange={(e) =>
                      setEditData({ ...editData, location: e.target.value })
                    }
                    className="w-full px-2 py-1 rounded border border-border bg-background text-sm"
                  />
                ) : (
                  <div className="text-sm flex items-center gap-1">
                    <MapPin className="w-3 h-3 text-muted-foreground" />
                    {item.location}
                  </div>
                )}
              </div>
              <div>
                <div className="text-xs text-muted-foreground">Quantity</div>
                <div className="text-sm">{item.quantity}</div>
              </div>
            </div>

            {/* eBay Info */}
            {item.listing_status === 'listed' && item.ebay_url && (
              <div className="mt-3 pt-3 border-t border-border">
                <div className="flex items-center gap-4">
                  {item.listed_price && (
                    <div>
                      <div className="text-xs text-muted-foreground">
                        Listed Price
                      </div>
                      <div className="text-sm font-medium">
                        {formatCurrency(item.listed_price)}
                      </div>
                    </div>
                  )}
                  <a
                    href={item.ebay_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-sm text-primary hover:underline"
                  >
                    <ExternalLink className="w-3 h-3" />
                    View on eBay
                  </a>
                </div>
              </div>
            )}

            {/* Sold Info */}
            {item.listing_status === 'sold' && (
              <div className="mt-3 pt-3 border-t border-border">
                <div className="grid grid-cols-3 gap-4">
                  {item.sold_price && (
                    <div>
                      <div className="text-xs text-muted-foreground">
                        Sold Price
                      </div>
                      <div className="text-sm font-medium text-green-400">
                        {formatCurrency(item.sold_price)}
                      </div>
                    </div>
                  )}
                  {item.sold_date && (
                    <div>
                      <div className="text-xs text-muted-foreground">
                        Sold Date
                      </div>
                      <div className="text-sm">{formatDate(item.sold_date)}</div>
                    </div>
                  )}
                  {item.sold_platform && (
                    <div>
                      <div className="text-xs text-muted-foreground">
                        Platform
                      </div>
                      <div className="text-sm capitalize">
                        {item.sold_platform}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>

          {/* Notes */}
          {(item.notes || isEditing) && (
            <div className="rounded-lg border border-border bg-card p-4">
              <h2 className="text-sm font-semibold mb-3">Notes</h2>
              {isEditing ? (
                <textarea
                  value={editData.notes ?? item.notes ?? ''}
                  onChange={(e) =>
                    setEditData({
                      ...editData,
                      notes: e.target.value || undefined,
                    })
                  }
                  rows={4}
                  className="w-full px-2 py-1 rounded border border-border bg-background text-sm resize-none"
                />
              ) : (
                <div className="text-sm text-muted-foreground whitespace-pre-wrap">
                  {item.notes}
                </div>
              )}
            </div>
          )}

          {/* Tags */}
          {item.tags && item.tags.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {item.tags.map((tag) => (
                <span
                  key={tag}
                  className="px-2 py-1 rounded-full bg-secondary text-xs"
                >
                  {tag}
                </span>
              ))}
            </div>
          )}

          {/* Timestamps */}
          <div className="text-xs text-muted-foreground flex items-center gap-4">
            <span>Created: {formatDate(item.created_at)}</span>
            <span>Updated: {formatDate(item.updated_at)}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

function StatusBadge({ status }: { status: ListingStatus }) {
  const colors: Record<ListingStatus, string> = {
    not_listed: 'bg-gray-500/20 text-gray-400',
    pending_review: 'bg-yellow-500/20 text-yellow-400',
    listed: 'bg-green-500/20 text-green-400',
    sold: 'bg-blue-500/20 text-blue-400',
    keeping: 'bg-purple-500/20 text-purple-400',
  };

  const labels: Record<ListingStatus, string> = {
    not_listed: 'Not Listed',
    pending_review: 'Pending Review',
    listed: 'Listed',
    sold: 'Sold',
    keeping: 'Keeping',
  };

  return (
    <span className={cn('px-2 py-0.5 rounded text-xs', colors[status])}>
      {labels[status]}
    </span>
  );
}

function ConditionBadge({ condition }: { condition: ItemCondition }) {
  const colors: Record<ItemCondition, string> = {
    new: 'bg-green-500/20 text-green-400',
    like_new: 'bg-cyan-500/20 text-cyan-400',
    good: 'bg-blue-500/20 text-blue-400',
    fair: 'bg-yellow-500/20 text-yellow-400',
    poor: 'bg-orange-500/20 text-orange-400',
    for_parts: 'bg-red-500/20 text-red-400',
  };

  const labels: Record<ItemCondition, string> = {
    new: 'New',
    like_new: 'Like New',
    good: 'Good',
    fair: 'Fair',
    poor: 'Poor',
    for_parts: 'For Parts',
  };

  return (
    <span className={cn('px-2 py-0.5 rounded text-xs', colors[condition])}>
      {labels[condition]}
    </span>
  );
}
