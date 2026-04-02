// Source: https://github.com/hyprwm/hyprland-plugins/blob/main/hyprbars/barDeco.cpp
// Extracted text rendering and decoration draw methods — reference for custom label plugin.

// === renderText: Generic text → OpenGL texture via Cairo/Pango ===
void CHyprBar::renderText(SP<CTexture> out, const std::string& text, const CHyprColor& color,
                          const Vector2D& bufferSize, const float scale, const int fontSize) {
    const auto CAIROSURFACE = cairo_image_surface_create(CAIRO_FORMAT_ARGB32, bufferSize.x, bufferSize.y);
    const auto CAIRO = cairo_create(CAIROSURFACE);

    // Clear
    cairo_save(CAIRO);
    cairo_set_operator(CAIRO, CAIRO_OPERATOR_CLEAR);
    cairo_paint(CAIRO);
    cairo_restore(CAIRO);

    // Pango text layout
    PangoLayout* layout = pango_cairo_create_layout(CAIRO);
    pango_layout_set_text(layout, text.c_str(), -1);

    PangoFontDescription* fontDesc = pango_font_description_from_string("sans");
    pango_font_description_set_size(fontDesc, fontSize * scale * PANGO_SCALE);
    pango_layout_set_font_description(layout, fontDesc);
    pango_font_description_free(fontDesc);

    const int maxWidth = bufferSize.x;
    pango_layout_set_width(layout, maxWidth * PANGO_SCALE);
    pango_layout_set_ellipsize(layout, PANGO_ELLIPSIZE_NONE);

    cairo_set_source_rgba(CAIRO, color.r, color.g, color.b, color.a);

    // Center text in buffer
    PangoRectangle ink_rect, logical_rect;
    pango_layout_get_extents(layout, &ink_rect, &logical_rect);
    const double xOffset = (bufferSize.x / 2.0 - ink_rect.width / PANGO_SCALE / 2.0);
    const double yOffset = (bufferSize.y / 2.0 - logical_rect.height / PANGO_SCALE / 2.0);

    cairo_move_to(CAIRO, xOffset, yOffset);
    pango_cairo_show_layout(CAIRO, layout);
    g_object_unref(layout);
    cairo_surface_flush(CAIROSURFACE);

    // Cairo surface → OpenGL texture
    const auto DATA = cairo_image_surface_get_data(CAIROSURFACE);
    out->allocate();
    glBindTexture(GL_TEXTURE_2D, out->m_texID);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST);
#ifndef GLES2
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_SWIZZLE_R, GL_BLUE);
    glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_SWIZZLE_B, GL_RED);
#endif
    glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, bufferSize.x, bufferSize.y, 0, GL_RGBA, GL_UNSIGNED_BYTE, DATA);

    cairo_destroy(CAIRO);
    cairo_surface_destroy(CAIROSURFACE);
}

// === draw(): Queue render pass element (called each frame) ===
void CHyprBar::draw(PHLMONITOR pMonitor, const float& a) {
    if (m_hidden || !validMapped(m_pWindow))
        return;

    auto data = CBarPassElement::SBarData{this, a};
    g_pHyprRenderer->m_renderPass.add(makeUnique<CBarPassElement>(data));
}

// === renderPass(): Actual OpenGL rendering (called by pass element) ===
void CHyprBar::renderPass(PHLMONITOR pMonitor, const float& a) {
    const auto PWINDOW = m_pWindow.lock();
    const auto PWORKSPACE = PWINDOW->m_workspace;
    const auto WORKSPACEOFFSET = PWORKSPACE && !PWINDOW->m_pinned ? PWORKSPACE->m_renderOffset->value() : Vector2D();
    const auto DECOBOX = assignedBoxGlobal();
    const auto BARBUF = DECOBOX.size() * pMonitor->m_scale;

    // Background rectangle
    CBox titleBarBox = {DECOBOX.x - pMonitor->m_position.x, DECOBOX.y - pMonitor->m_position.y,
                        DECOBOX.w, DECOBOX.h};
    titleBarBox.translate(PWINDOW->m_floatingOffset).scale(pMonitor->m_scale).round();

    CHyprColor color = {0, 0, 0, 0.5}; // semi-transparent
    g_pHyprOpenGL->renderRect(titleBarBox, color, {});

    // Text texture
    if (m_szLastTitle != PWINDOW->m_title || m_pTextTex->m_texID == 0) {
        m_szLastTitle = PWINDOW->m_title;
        renderBarTitle(BARBUF, pMonitor->m_scale);
    }

    CBox textBox = {titleBarBox.x, titleBarBox.y, (int)BARBUF.x, (int)BARBUF.y};
    g_pHyprOpenGL->renderTexture(m_pTextTex, textBox, {.a = a});
}

// === getPositioningInfo(): Declare decoration geometry ===
SDecorationPositioningInfo CHyprBar::getPositioningInfo() {
    SDecorationPositioningInfo info;
    // STICKY = reserves space, ABSOLUTE = overlay without layout impact
    info.policy = DECORATION_POSITION_STICKY;
    info.edges = DECORATION_EDGE_TOP;
    info.priority = 5000;
    info.reserved = true;
    info.desiredExtents = {{0, barHeight}, {0, 0}};
    return info;
}

// === For a label that OVERLAYS the window (doesn't push it down): ===
// info.policy = DECORATION_POSITION_ABSOLUTE;
// info.edges = DECORATION_EDGE_BOTTOM; // or TOP
// info.reserved = false;

// === getDecorationLayer(): Where in z-order ===
eDecorationLayer CHyprBar::getDecorationLayer() {
    // DECORATION_LAYER_UNDER = behind window content
    // DECORATION_LAYER_OVER = on top of window content
    return DECORATION_LAYER_UNDER;
}
