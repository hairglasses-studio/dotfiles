# tailscale_unraid_plugin

*Small utility consolidated from standalone repository*

---

# 🔮 Tailscale Unraid Dashboard

Advanced monitoring dashboard for Unraid services and Tailscale network nodes with psychedelic aesthetics and sacred geometry themes.

## ✨ Features

- **Real-time Monitoring**: Live updates of system metrics, Docker containers, and Tailscale nodes
- **Beautiful UI**: Psychedelic theme with sacred geometry animations and glowing effects
- **Comprehensive Metrics**: CPU, memory, disk usage, network I/O, and container statistics
- **Tailscale Integration**: Monitor all nodes in your Tailscale network
- **Responsive Design**: Works on desktop and mobile devices

## 🚀 Installation

1. Download the plugin files to your Unraid server
2. Run the installation script:
   ```bash
   chmod +x install.sh
   ./install.sh
   ```
3. Set your Tailscale API key in `/etc/systemd/system/tailscale-dashboard.service`
4. Access the dashboard at `http://your-unraid-ip:8080`

##  Configuration

### Tailscale API Key
Get your API key from the [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys) and update the service file:

```bash
sudo nano /etc/systemd/system/tailscale-dashboard.service
```

Replace `your_api_key_here` with your actual API key.

### Customization
- Modify `templates/dashboard.html` for visual changes
- Edit `app.py` for additional metrics
- Add new monitoring endpoints in the Flask app

## 📊 Metrics Tracked

### System Metrics
- CPU usage percentage
- Memory usage and available RAM
- Disk usage across all mounted volumes
- Network I/O statistics
- System load averages

### Docker Services
- Container status (running/stopped)
- CPU and memory usage per container
- Network traffic per container
- Container uptime

### Tailscale Network
- Node online/offline status
- IP addresses and last seen timestamps
- Operating system information
- Network topology visualization

##  Aesthetic Features

- **Sacred Geometry**: Animated geometric patterns in the background
- **Glowing Effects**: Text and border animations
- **Gradient Backgrounds**: Deep space-inspired color schemes
- **Smooth Animations**: Real-time metric updates with smooth transitions
- **Responsive Design**: Adapts to different screen sizes

## 🔧 Development

### Prerequisites
- Python 3.8+
- Docker API access
- Tailscale API key
- Unraid 6.8.0+

### Local Development
```bash
pip install -r requirements.txt
python app.py
```

### Adding New Metrics
1. Add metric collection in `TailscaleMonitor` class
2. Update the API endpoint in `app.py`
3. Add visualization in `dashboard.html`

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📝 License

MIT License - see LICENSE file for details

## 🆘 Support

- GitHub Issues: [Create an issue](https://github.com/hairglasses/tailscale_unraid_plugin/issues)
- Documentation: Check the wiki for detailed guides
- Community: Join our Discord for discussions

---

*Built with ❤️ for the Unraid and Tailscale communities*


## Files from Original Repository

- `requirements.txt` - Documentation
- `plugin.cfg`
- `install.sh` - Script/executable
- `README.md` - Documentation
- `tailscale-dashboard.plg`
- `create_release.sh` - Script/executable
- `test_plugin.sh` - Script/executable
- `app.py` - Script/executable
- `package_plugin.sh` - Script/executable
- `build_plugin.sh` - Script/executable
- `api_endpoints.py` - Script/executable
- `create_assets.sh` - Script/executable
- `templates/dashboard.html`
- `templates/settings.html`
- `.git/config`
- `.git/HEAD`
- `.git/description`
- `.git/index`
- `.git/packed-refs`
- `.git/COMMIT_EDITMSG`
- `.git/FETCH_HEAD`
- `.git/info/exclude`
- `.git/logs/HEAD`
- `.git/hooks/commit-msg.sample`
- `.git/hooks/pre-rebase.sample`
- `.git/hooks/pre-commit.sample`
- `.git/hooks/applypatch-msg.sample`
- `.git/hooks/fsmonitor-watchman.sample`
- `.git/hooks/pre-receive.sample`
- `.git/hooks/prepare-commit-msg.sample`
- `.git/hooks/post-update.sample`
- `.git/hooks/pre-merge-commit.sample`
- `.git/hooks/pre-applypatch.sample`
- `.git/hooks/pre-push.sample`
- `.git/hooks/update.sample`
- `.git/hooks/push-to-checkout.sample`
- `.git/objects/9d/f72b2385c1d31fa1b5bb27fb01c40c07851759`
- `.git/objects/pack/pack-095d7646ff11e3e35e521f2660f57c31e9227266.idx`
- `.git/objects/pack/pack-095d7646ff11e3e35e521f2660f57c31e9227266.pack`
- `.git/objects/b1/b7497eea24ef40087f8164b2082d1f4e6c84a5`
- `.git/objects/dc/65176903e88ad951f932680440cb0a03a43779`
- `.git/logs/refs/heads/main`
- `.git/logs/refs/remotes/origin/HEAD`
- `.git/logs/refs/remotes/origin/main`
- `.git/refs/heads/main`
- `.git/refs/remotes/origin/HEAD`
- `.git/refs/remotes/origin/main`

---

*Repository consolidated on 2025-09-23 - originally located at `tailscale_unraid_plugin/`*

**Note:** This utility was small enough to be documented here rather than maintained as a separate repository. 
If active development resumes, consider recreating as a standalone repository.
