
#修改默认SSH端口
sudo sed -i 's/#Port 22/Port 65532/g' /etc/ssh/sshd_config && sudo systemctl restart sshd
sudo yum install -y iptables-services
sudo iptables -P INPUT ACCEPT
sudo iptables -P FORWARD ACCEPT
sudo iptables -P OUTPUT ACCEPT
sudo iptables -F INPUT
sudo iptables -F FORWARD
sudo iptables -F OUTPUT
sudo iptables -F
sudo iptables -X
sudo iptables -Z
sudo  iptables -t mangle -F
sudo  iptables -t mangle -N DIVERT
sudo  iptables -t mangle -A DIVERT -j MARK --set-mark 1
sudo  iptables -t mangle -A DIVERT -j ACCEPT
sudo  iptables -t mangle -I PREROUTING -p tcp -m socket -j DIVERT
sudo  iptables -t mangle -I PREROUTING -p udp -m socket -j DIVERT
sudo  iptables -t mangle -A PREROUTING -i $(ip -o -4 route show to default | awk '{print $5}') -p tcp -d  $(hostname -I | awk '{print $1}')  --dport 0:12344 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p tcp -d  $(hostname -I | awk '{print $1}') --dport 12346:65520 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p udp -d  $(hostname -I | awk '{print $1}') --dport 0:12344 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345
sudo  iptables -t mangle -A PREROUTING -i  $(ip -o -4 route show to default | awk '{print $5}') -p udp -d  $(hostname -I | awk '{print $1}') --dport 12346:65520 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip   $(hostname -I | awk '{print $1}')  --on-port 12345

sudo ip6tables -P INPUT ACCEPT
sudo ip6tables -P FORWARD ACCEPT
sudo ip6tables -P OUTPUT ACCEPT
sudo ip6tables -F INPUT
sudo ip6tables -F FORWARD
sudo ip6tables -F OUTPUT
sudo ip6tables -F
sudo ip6tables -X
sudo ip6tables -Z
sudo  ip6tables -t mangle -F
sudo  ip6tables -t mangle -N DIVERT
sudo  ip6tables -t mangle -A DIVERT -j MARK --set-mark 1
sudo  ip6tables -t mangle -A DIVERT -j ACCEPT
sudo  ip6tables -t mangle -I PREROUTING -p tcp -m socket -j DIVERT
sudo  ip6tables -t mangle -I PREROUTING -p udp -m socket -j DIVERT
sudo ip6tables -t mangle -A PREROUTING -i $(ip -o -6 route show to default | awk '{print $5}') -p udp -d $(hostname -I | awk '{print $2}') --dport 0:65535 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip $(hostname -I | awk '{print $2}') --on-port 12345
sudo ip6tables -t mangle -A PREROUTING -i $(ip -o -6 route show to default | awk '{print $5}') -p tcp -d $(hostname -I | awk '{print $2}') --dport 0:65535 -j TPROXY --tproxy-mark 0x1/0x1 --on-ip $(hostname -I | awk '{print $2}') --on-port 12345

sudo service iptables save
sudo systemctl start iptables
sudo systemctl enable iptables

sudo service ip6tables save
sudo systemctl start ip6tables
sudo systemctl enable ip6tables
